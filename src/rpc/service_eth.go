package rpc

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	txtypes "github.com/ethereum/go-ethereum/core/types"
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

var EpochComputer = shcrypto.ComputeEpochID

func ComputeIdentity(prefix []byte, sender common.Address) *shcrypto.EpochID {
	bytes := append(prefix, sender.Bytes()...)
	return EpochComputer(bytes)
}

func ComputeSlotFunc(blockTimestamp uint64) (*uint64, error) {
	if blockTimestamp < uint64(GENESIS_TIME) {
		return nil, errors.New("Slot computation error")
	}
	if (blockTimestamp-uint64(GENESIS_TIME))%uint64(SECONDS_PER_SLOT) != 0 {
		return nil, errors.New("Slot computation error")
	}

	slot := (blockTimestamp - uint64(GENESIS_TIME)) / uint64(SECONDS_PER_SLOT)

	return &slot, nil
}

var ComputeSlot = ComputeSlotFunc

type EthService struct {
	processor Processor
}

func (s *EthService) InjectProcessor(p Processor) {
	s.processor = p
}

func (s *EthService) Name() string {
	return "eth"
}

func (service *EthService) SendTransaction(ctx context.Context, tx *txtypes.Transaction) (*common.Hash, error) {
	ts := txtypes.Transactions{tx}
	buf := new(bytes.Buffer)
	ts.EncodeIndex(0, buf)
	rawTx := hexutil.Encode(buf.Bytes())

	return service.SendRawTransaction(ctx, rawTx)
}

func (service *EthService) SendRawTransaction(ctx context.Context, s string) (*common.Hash, error) {
	block, err := service.processor.Client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	slot, err := ComputeSlot(block.Header().Time)
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
	signer := txtypes.NewLondonSigner(tx.ChainId())
	fromAddress, err := signer.Sender(tx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	accountNonce, err := service.processor.Client.NonceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	if accountNonce != tx.Nonce() {
		return nil, &EncodingError{StatusCode: -32000, Err: errors.New("nonce is not correct")}
	}

	accountBalance, err := service.processor.Client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	if accountBalance.Cmp(tx.Cost()) == -1 {
		return nil, &EncodingError{StatusCode: -32000, Err: errors.New("gas cost is higher")}
	}

	eon, err := service.processor.KeyperSetManagerContract.GetKeyperSetIndexBySlot(nil, *slot+uint64(service.processor.KeyperSetChangeLookAhead))
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	eonKeyBytes, err := service.processor.KeyBroadcastContract.GetEonKey(nil, eon)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	eonKey := shcrypto.EonPublicKey{}
	if err := eonKey.Unmarshal(eonKeyBytes); err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	identityPrefix, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}
	identity := ComputeIdentity(identityPrefix[:], fromAddress)

	encryptedTx := shcrypto.Encrypt(b, &eonKey, identity, sigma)

	chainId, err := service.processor.Client.ChainID(ctx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}

	newSigner, err := bind.NewKeyedTransactorWithChainID(service.processor.SigningKey, chainId)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	opts := bind.TransactOpts{
		From:   *service.processor.SigningAddress,
		Signer: newSigner.Signer,
		NoSend: true,
	}

	submitTx, err := service.processor.SequencerContract.SubmitEncryptedTransaction(&opts, eon, identityPrefix, encryptedTx.Marshal(), new(big.Int).SetUint64(tx.Gas()))
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}

	signerBalance, err := service.processor.Client.BalanceAt(ctx, *service.processor.SigningAddress, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}

	if signerBalance.Cmp(submitTx.Cost()) == -1 {
		return nil, &EncodingError{StatusCode: -32003, Err: errors.New("signer is lacking funds")}
	}

	opts.NoSend = false
	opts.Value = submitTx.Cost()

	submitTx, err = service.processor.SequencerContract.SubmitEncryptedTransaction(&opts, eon, identityPrefix, encryptedTx.Marshal(), new(big.Int).SetUint64(tx.Gas()))
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}

	_, err = bind.WaitMined(ctx, service.processor.Client, submitTx)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32603, Err: err}
	}

	txHash := tx.Hash()
	return &txHash, nil
}
