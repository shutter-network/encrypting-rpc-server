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

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/shutter-network/encrypting-rpc-server/cache"
	"github.com/shutter-network/encrypting-rpc-server/db"
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
					metrics.ErrorReturnedGauge.Dec()
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
		return nil, returnError(-32602, err)
	}

	b, err := hexutil.Decode(s)
	if err != nil {
		return nil, returnError(-32602, err)
	}
	tx := new(txtypes.Transaction)

	if err := tx.UnmarshalBinary(b); err != nil {
		return nil, returnError(-32602, err)
	}

	txHash := tx.Hash()
	fromAddress, err := utils.SenderAddress(tx)
	if err != nil {
		return nil, returnError(-32602, err)
	}

	accountNonce, err := service.Processor.Client.NonceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, returnError(-32602, err)
	}

	if accountNonce > tx.Nonce() {
		return nil, returnError(-32000, errors.New("nonce is not correct"))
	}

	accountBalance, err := service.Processor.Client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, returnError(-32602, err)
	}

	if accountBalance.Cmp(tx.Cost()) == -1 {
		return nil, returnError(-32000, errors.New("gas cost is higher"))
	}

	intrinsicGas, err := CalculateIntrinsicGas(tx)
	if err != nil {
		return nil, returnError(-32602, errors.New("error calculating the intrinsic gas: "+err.Error()))
	}

	if tx.Gas() < intrinsicGas {
		return nil, returnError(-32602, errors.New("gas limit below the intrinsic gas limit "+
			""+strconv.FormatUint(intrinsicGas, 10)))
	}

	if tx.Gas() > service.Config.EncryptedGasLimit {
		return nil, returnError(-32000, errors.New("gas limit exceeds encrypted gas limit "+
			"(max gas limit allowed per shutterized block)"))
	}

	if utils.IsCancellationTransaction(tx, fromAddress) {
		utils.Logger.Info().Msg("Detected cancellation transaction, forwarding to backend")

		backendClient, err := rpc.Dial(service.Config.BackendURL.String())
		if err != nil {
			utils.Logger.Err(err).Msg("Failed to connect to backend")
			return nil, returnError(-32603, err)
		}

		err = backendClient.CallContext(ctx, &txHash, "eth_sendRawTransaction", s)
		if err != nil {
			utils.Logger.Err(err).Msg("Failed to send cancel transaction to backend")
			return nil, returnError(-32602, err)
		}

		metrics.CancellationTxGauge.WithLabelValues(txHash.String()).Inc()
		utils.Logger.Info().Msg("Transaction forwarded with hash: " + txHash.Hex())

		service.Processor.Db.InsertNewTx(db.TransactionDetails{
			Address:        fromAddress.String(),
			Nonce:          tx.Nonce(),
			TxHash:         txHash.String(),
			SubmissionTime: time.Now().Unix(),
			IsCancel:       true,
		})

		_ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(service.Config.WaitMinedInterval)*10*time.Second)
		go service.WaitTillMined(_ctx, cancelFunc, tx, service.Config.WaitMinedInterval)
		return &txHash, nil
	}

	statuses, err := service.Cache.ProcessTxEntry(tx, time.Now().Unix())
	if err != nil {
		utils.Logger.Err(err).Msg("Failed to update the cache.")
		return nil, returnError(-32603, err)
	}

	if !statuses.SendStatus {
		utils.Logger.Info().Hex("Tx hash", txHash.Bytes()).Msg("Transaction delayed")
		if statuses.UpdateStatus { // this is the same tx, just requested more than once so we do not add it to db
			service.Processor.Db.InsertNewTx(db.TransactionDetails{
				Address: fromAddress.String(),
				Nonce:   tx.Nonce(),
				TxHash:  txHash.String(),
			})
		}
		return &txHash, nil
	}

	submitTx, err := service.ProcessTransaction(tx, ctx, service, blockNumber, b)
	if err != nil {
		return nil, returnError(-32603, err)
	}
	utils.Logger.Info().Hex("Incoming tx hash", txHash.Bytes()).Hex("Encrypted tx hash", submitTx.Hash().Bytes()).Msg("Transaction sent")

	service.Processor.Db.InsertNewTx(db.TransactionDetails{
		Address:         fromAddress.String(),
		Nonce:           tx.Nonce(),
		TxHash:          txHash.String(),
		EncryptedTxHash: submitTx.Hash().String(),
		SubmissionTime:  time.Now().Unix(),
	})

	_ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(service.Config.WaitMinedInterval)*10*time.Second)
	go service.WaitTillMined(_ctx, cancelFunc, tx, service.Config.WaitMinedInterval)

	metrics.RequestedGasLimit.WithLabelValues(submitTx.Hash().String()).Observe(float64(tx.Gas()))
	metrics.TotalRequestDuration.WithLabelValues(submitTx.Hash().String(), txHash.String()).Observe(float64(time.Since(timeBefore).Seconds()))

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

	encryptionDuration := time.Since(timeBefore).Seconds()

	opts := bind.TransactOpts{
		From:   *service.Processor.SigningAddress,
		Signer: newSigner.Signer,
	}

	opts.Value = big.NewInt(0).Sub(tx.Cost(), tx.Value())

	submitTx, err := service.Processor.SequencerContract.SubmitEncryptedTransaction(&opts, eon, identityPrefix, encryptedTx.Marshal(), new(big.Int).SetUint64(tx.Gas()))
	if err != nil {
		return nil, err
	}
	metrics.EncryptionDuration.WithLabelValues(submitTx.Hash().String()).Observe(float64(encryptionDuration))

	return submitTx, nil
}

func (s *EthService) WaitTillMined(ctx context.Context, cancelFunc context.CancelFunc, tx *types.Transaction, waitMinedInterval int) {
	key := tx.Hash().String()
	value := s.Cache.WaitingForReceiptCache[key]

	if !value {
		queryTicker := time.NewTicker(time.Duration(waitMinedInterval) * time.Second)
		defer queryTicker.Stop()
		s.Cache.WaitingForReceiptCache[key] = true
		utils.Logger.Info().Msgf("New tx recorded to check for inclusion | txHash: %s", tx.Hash().String())
		for {
			if s.Cache.WaitingForReceiptCache[key] {
				receipt, err := s.Processor.Client.TransactionReceipt(ctx, tx.Hash())
				if err == nil {
					delete(s.Cache.WaitingForReceiptCache, key)
					block, err := s.Processor.Client.BlockByHash(ctx, receipt.BlockHash)
					if err != nil {
						utils.Logger.Debug().Msgf("Error getting block | blockHash: %s", receipt.BlockHash.String())
						cancelFunc()

					}

					s.Processor.Db.FinaliseTx(db.TransactionDetails{
						TxHash:        tx.Hash().String(),
						InclusionTime: block.Time(),
					})
					cancelFunc()
				}

				if errors.Is(err, ethereum.NotFound) {
					utils.Logger.Debug().Msgf("Transaction not yet mined | txHash: %s", tx.Hash().String())
				} else {
					delete(s.Cache.WaitingForReceiptCache, key)
					utils.Logger.Debug().Msgf("receipt retrieval failed | txHash: %s | err: %W", tx.Hash().String(), err)
					cancelFunc()
				}

			} else {
				// return if tx is explicitely set to not check
				cancelFunc()
			}
			// Wait for the next round.
			select {
			case <-ctx.Done():
				// deleting cache here as we have stopped waiting for tx inclusion
				delete(s.Cache.WaitingForReceiptCache, key)
				return
			case <-queryTicker.C:
			}
		}
	}
}

func (p *Processor) MonitorBalance(ctx context.Context, delayInSeconds int) {
	timer := time.NewTicker(time.Duration(delayInSeconds) * time.Second)

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info().Msg("Stopping because context is done.")
			return

		case <-timer.C:
			balance, err := p.Client.BalanceAt(ctx, *p.SigningAddress, nil)
			if err != nil {
				utils.Logger.Err(err).Msg("Failed to get balance")
				continue
			}

			// Convert balance from Wei to Ether
			ethValue := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
			balanceInFloat, _ := ethValue.Float64()

			metrics.ERPCBalance.Set(balanceInFloat)
		}
	}
}

func returnError(status int, msg error) *EncodingError {
	metrics.ErrorReturnedGauge.Inc()
	return &EncodingError{
		StatusCode: status,
		Err:        msg,
	}
}
