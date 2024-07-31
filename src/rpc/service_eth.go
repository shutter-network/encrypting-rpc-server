package rpc

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shutter-network/encrypting-rpc-server/cache"
	"github.com/shutter-network/encrypting-rpc-server/metrics"
	"github.com/shutter-network/encrypting-rpc-server/utils"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/shutter/shlib/shcrypto"
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
}

func (s *EthService) Init(processor Processor, config Config) {
	s.Processor = processor
	s.Config = config
	s.Cache = cache.NewCache(int64(config.DelayInSeconds))
}

func (s *EthService) Name() string {
	return "eth"
}

func (s *EthService) SendTimeEvents(ctx context.Context, delayInSeconds int) {
	timer := time.NewTicker(time.Duration(delayInSeconds) * time.Second)

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info().Msg("Stopping because context is done.")
			return

		case tickTime := <-timer.C:
			newTime := tickTime.Unix()
			utils.Logger.Debug().Msgf("Received timer event | Unix time = [%d] | Time = [%v]",
				newTime, time.Unix(newTime, 0))

			s.NewTimeEvent(ctx, newTime)
		}
	}
}

func (s *EthService) NewTimeEvent(ctx context.Context, newTime int64) {
	utils.Logger.Info().Msg(fmt.Sprintf("Received new time event: %d", newTime))
	for key, info := range s.Cache.Data {
		if info.CachedTime+s.Cache.DelayFactor <= newTime {
			utils.Logger.Debug().Msgf("Deleting entry at key [%s]", key)
			delete(s.Cache.Data, key)

			if info.Tx != nil {
				utils.Logger.Debug().Msgf("Sending transaction [%s]", info.Tx.Hash().Hex())
				rawTxBytes, err := info.Tx.MarshalBinary()
				if err != nil {
					utils.Logger.Error().Err(err).Msg("Failed to marshal data")
				}

				rawTx := "0x" + common.Bytes2Hex(rawTxBytes)
				txHash, err := s.SendRawTransaction(ctx, rawTx)

				if err != nil {
					utils.Logger.Error().Err(err).Msgf("Failed to send transaction.")
					continue
				}

				utils.Logger.Info().Msg("Transaction sent internally: " + txHash.Hex())
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
	timeBefore := time.Now()
	if service.ProcessTransaction == nil {
		service.ProcessTransaction = DefaultProcessTransaction
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
	fromAddress, err := utils.SenderAddress(tx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	accountNonce, err := service.Processor.Client.NonceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	if accountNonce > tx.Nonce() {
		return nil, &EncodingError{StatusCode: -32000, Err: errors.New("nonce is not correct")}
	}

	accountBalance, err := service.Processor.Client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	if accountBalance.Cmp(tx.Cost()) == -1 {
		return nil, &EncodingError{StatusCode: -32000, Err: errors.New("gas cost is higher")}
	}

	intrinsicGas, err := CalculateIntrinsicGas(tx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: errors.New("error calculating the intrinsic gas: " + err.Error())}
	}

	if tx.Gas() < intrinsicGas {
		return nil, &EncodingError{StatusCode: -32602, Err: errors.New("gas limit below the intrinsic gas limit " +
			"" + strconv.FormatUint(intrinsicGas, 10))}
	}

	if tx.Gas() > service.Config.EncryptedGasLimit {
		return nil, &EncodingError{StatusCode: -32000, Err: errors.New("gas limit exceeds encrypted gas limit " +
			"(max gas limit allowed per shutterized block)")}
	}

	if utils.IsCancellationTransaction(tx, fromAddress) {
		utils.Logger.Info().Msg("Detected cancellation transaction, forwarding to backend")

		backendClient, err := rpc.Dial(service.Config.BackendURL.String())
		if err != nil {
			utils.Logger.Err(err).Msg("Failed to connect to backend")
			return nil, &EncodingError{StatusCode: -32603, Err: err}
		}

		err = backendClient.CallContext(ctx, &txHash, "eth_sendRawTransaction", s)
		if err != nil {
			utils.Logger.Err(err).Msg("Failed to send cancel transaction to backend")
			return nil, &EncodingError{StatusCode: -32602, Err: err}
		}

		utils.Logger.Info().Msg("Transaction forwarded with hash: " + txHash.Hex())
		return &txHash, nil
	}

	sendStatus, err := service.Cache.ProcessTxEntry(tx, time.Now().Unix())
	if err != nil {
		utils.Logger.Err(err).Msg("Failed to update the cache.")
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}

	if !sendStatus {
		utils.Logger.Info().Hex("Tx hash", txHash.Bytes()).Msg("Transaction delayed")
		return &txHash, nil
	}

	submitTx, err := service.ProcessTransaction(tx, ctx, service, blockNumber, b)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}
	utils.Logger.Info().Hex("Incoming tx hash", txHash.Bytes()).Hex("Encrypted tx hash", submitTx.Hash().Bytes()).Msg("Transaction sent")

	metrics.MetricsRequestedGasLimit.WithLabelValues(txHash.String()).Observe(float64(tx.Gas()))
	metrics.MetricsTotalRequestDuration.WithLabelValues(txHash.String()).Observe(float64(time.Since(timeBefore)))
	return &txHash, nil
}

var DefaultProcessTransaction = func(tx *txtypes.Transaction, ctx context.Context, service *EthService, blockNumber uint64, b []byte) (*txtypes.Transaction, error) {
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

	timeBefore := time.Now()

	identityPrefix, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	identity := ComputeIdentity(identityPrefix[:], newSigner.From)
	encryptedTx := shcrypto.Encrypt(b, eonKey, identity, sigma)

	metrics.MetricsEncryptionDuration.WithLabelValues(tx.Hash().String()).Observe(float64(time.Since(timeBefore).Seconds()))

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
