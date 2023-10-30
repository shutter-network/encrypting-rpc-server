package rpc

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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

func ComputeIdentity(prefix []byte, sender common.Address) *shcrypto.EpochID {
	bytes := append(prefix, sender.Bytes()...)
	return shcrypto.ComputeEpochID(bytes)
}

func ComputeSlot(blockTimestamp uint64) (*uint64, error) {
	if blockTimestamp < uint64(GENESIS_TIME) {
		return nil, errors.New("Slot computation error")
	}
	if (blockTimestamp-uint64(GENESIS_TIME))%uint64(SECONDS_PER_SLOT) != 0 {
		return nil, errors.New("Slot computation error")
	}

	slot := (blockTimestamp - uint64(GENESIS_TIME)) / uint64(SECONDS_PER_SLOT)

	return &slot, nil
}

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
	rawSignature := new(bytes.Buffer)
	err := tx.EncodeRLP(rawSignature)
	if err != nil {
		return nil, err
	}

	return service.SendRawTransaction(ctx, hex.EncodeToString(rawSignature.Bytes()))
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

	b, err := hex.DecodeString(s)
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
		return nil, &EncodingError{StatusCode: -32000, Err: err}
	}

	accountBalance, err := service.processor.Client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		return nil, &EncodingError{StatusCode: -32602, Err: err}
	}

	if accountBalance.Cmp(tx.Cost()) == -1 {
		return nil, &EncodingError{StatusCode: -32000, Err: err}
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

	newSigner, err := bind.NewKeyedTransactorWithChainID(service.processor.SigningKey, tx.ChainId())
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
		return nil, &EncodingError{StatusCode: -32003, Err: err}
	}

	service.processor.Client.SendTransaction(ctx, submitTx)

	txHash := tx.Hash()
	return &txHash, nil
}
