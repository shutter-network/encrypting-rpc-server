package rpc

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shutter-network/encrypting-rpc-server/cache"
	"github.com/shutter-network/encrypting-rpc-server/requests"
	"github.com/shutter-network/encrypting-rpc-server/utils"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/shutter/shlib/shcrypto"
	"math/big"
)

var (
	GENESIS_TIME     = 1638993340
	SECONDS_PER_SLOT = 5
)

type EncodingError struct {
	StatusCode int
	Err        error
}

func (r *EncodingError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Err)
}

func ComputeIdentity(prefix []byte, sender common.Address) *shcrypto.EpochID {
	imageBytes := append(prefix, sender.Bytes()...)
	return shcrypto.ComputeEpochID(identitypreimage.IdentityPreimage(imageBytes).Bytes())
}

type EthService struct {
	Processor          Processor
	Config             Config
	Cache              *cache.Cache
	ProcessTransaction func(tx *txtypes.Transaction, ctx context.Context, service *EthService, blockNumber uint64, b []byte) (*txtypes.Transaction, error)
	WaitMinedFunc      func(ctx context.Context, backend bind.DeployBackend, tx *txtypes.Transaction) (*txtypes.Receipt, error)
}

func (s *EthService) InjectProcessor(p Processor) {
	s.Processor = p
}

func (s *EthService) AddConfig(config Config) {
	s.Config = config
	s.Cache = cache.NewCache(uint64(config.DelayFactor))
}

func (s *EthService) Name() string {
	return "eth"
}

func (s *EthService) NewBlock(ctx context.Context, blockNumber uint64) {
	utils.Logger.Info().Msg(fmt.Sprintf("Received blockNumber: %d", blockNumber))
	for key, info := range s.Cache.Data {
		if info.SentBlock+s.Cache.DelayFactor == blockNumber { // todo reorg issue? <=
			if info.Tx == nil {
				utils.Logger.Debug().Msgf("Deleting entry at key [%s]", key)
				delete(s.Cache.Data, key)
			} else {
				utils.Logger.Debug().Msgf("Sending transaction [%s] to the sequencer from block listener", info.Tx.Hash().Hex())
				rawTxBytes, err := info.Tx.MarshalBinary()
				if err != nil {
					utils.Logger.Error().Err(err).Msg("Failed to marshal data")
				}

				rawTx := "0x" + common.Bytes2Hex(rawTxBytes)

				txHash, err := s.SendRawTransaction(ctx, rawTx)
				if err != nil {
					utils.Logger.Error().Err(err).Msg("Failed to send transaction")
					continue
				}

				utils.Logger.Info().Msg("Transaction sent: " + txHash.Hex())
				info.Tx = nil
				info.SentBlock = blockNumber
				s.Cache.Data[key] = info
			}
		}
	}
}

func (service *EthService) SendTransaction(ctx context.Context, tx *txtypes.Transaction) (*common.Hash, error) {
	ts := txtypes.Transactions{tx}
	buf := new(bytes.Buffer)
	ts.EncodeIndex(0, buf)
	rawTx := hexutil.Encode(buf.Bytes())

	return service.SendRawTransaction(ctx, rawTx)
}

func (service *EthService) SendRawTransaction(ctx context.Context, s string) (*common.Hash, error) {
	if service.ProcessTransaction == nil {
		service.ProcessTransaction = DefaultProcessTransaction
	}

	if service.WaitMinedFunc == nil {
		service.WaitMinedFunc = DefaultWaitMined
	}

	blockNumber, err := service.Processor.Client.BlockNumber(ctx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	b, err := hexutil.Decode(s)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	tx := new(txtypes.Transaction)

	if err := tx.UnmarshalBinary(b); err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	txHash := tx.Hash()
	txFromAddress, err := utils.SenderAddress(tx)

	if utils.IsCancellationTransaction(tx, txFromAddress) {
		utils.Logger.Info().Msg("Detected cancellation transaction, sending it right away...")

		backendClient, err := rpc.Dial(service.Config.BackendURL.String())
		if err != nil {
			utils.Logger.Err(err).Msg("Failed to connect to the Ethereum client")
		}

		txHash := requests.SendTx(backendClient, s)
		utils.Logger.Info().Msg("Transaction forwarded with hash: " + txHash.Hex())
		return &txHash, nil
	}

	updated, err := service.Cache.UpdateEntry(tx, blockNumber)
	if err != nil {
		utils.Logger.Err(err).Msg("Failed to update the cache.")
		return nil, &EncodingError{StatusCode: -32602, Err: err} // todo check if necessary
	}

	if !updated {
		utils.Logger.Info().Hex("Tx hash", txHash.Bytes()).Msg("Transaction delayed")
		return &txHash, nil
	}

	submitTx, err := service.ProcessTransaction(tx, ctx, service, blockNumber, b)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}
	utils.Logger.Info().Hex("Incoming tx hash", txHash.Bytes()).Hex("Encrypted tx hash", submitTx.Hash().Bytes()).Msg("Transaction sent")

	_, err = service.WaitMinedFunc(ctx, service.Processor.Client, submitTx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}

	return &txHash, nil
}

var DefaultWaitMined = func(ctx context.Context, backend bind.DeployBackend, tx *txtypes.Transaction) (*txtypes.Receipt, error) {
	mined, err := bind.WaitMined(ctx, backend, tx)
	if err != nil {
		return nil, err
	}
	return mined, nil
}

var DefaultProcessTransaction = func(tx *txtypes.Transaction, ctx context.Context, service *EthService, blockNumber uint64, b []byte) (*txtypes.Transaction, error) {
	signer := txtypes.NewLondonSigner(tx.ChainId())
	fromAddress, err := signer.Sender(tx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	accountNonce, err := service.Processor.Client.NonceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	if accountNonce != tx.Nonce() {
		return nil, &EncodingError{StatusCode: -32000, Err: errors.New("nonce is not correct")}
	}

	accountBalance, err := service.Processor.Client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	if accountBalance.Cmp(tx.Cost()) == -1 {
		return nil, &EncodingError{StatusCode: -32000, Err: errors.New("gas cost is higher")}
	}

	eon, err := service.Processor.KeyperSetManagerContract.GetKeyperSetIndexByBlock(nil, blockNumber+uint64(service.Processor.KeyperSetChangeLookAhead))
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	eonKeyBytes, err := service.Processor.KeyBroadcastContract.GetEonKey(nil, eon)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	eonKey := &shcrypto.EonPublicKey{}
	if err := eonKey.Unmarshal(eonKeyBytes); err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	chainId, err := service.Processor.Client.ChainID(ctx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}

	newSigner, err := bind.NewKeyedTransactorWithChainID(service.Processor.SigningKey, chainId)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	identityPrefix, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	identity := ComputeIdentity(identityPrefix[:], newSigner.From)
	encryptedTx := shcrypto.Encrypt(b, eonKey, identity, sigma)

	opts := bind.TransactOpts{
		From:   *service.Processor.SigningAddress,
		Signer: newSigner.Signer,
	}

	opts.Value = big.NewInt(0).Sub(tx.Cost(), tx.Value())

	submitTx, err := service.Processor.SequencerContract.SubmitEncryptedTransaction(&opts, eon, identityPrefix, encryptedTx.Marshal(), new(big.Int).SetUint64(tx.Gas()))
	if err != nil {
		return nil, err
	}

	return submitTx, nil

}
