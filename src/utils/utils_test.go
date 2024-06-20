package utils

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"

	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestIsCancellationTransaction(t *testing.T) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("Failed to assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	anotherAddress := common.HexToAddress("0xC0058BdcC93EaA1afd468f06A26394E2d80c8f01")

	tests := []struct {
		name     string
		tx       *txtypes.Transaction
		expected bool
	}{
		{
			name:     "Cancellation transaction with nil To address",
			tx:       txtypes.NewTx(&txtypes.LegacyTx{Nonce: 0, To: nil, Value: big.NewInt(0)}),
			expected: true,
		},
		{
			name:     "Cancellation transaction with zero To address",
			tx:       txtypes.NewTx(&txtypes.LegacyTx{Nonce: 0, To: &zeroAddress, Value: big.NewInt(0)}),
			expected: true,
		},
		{
			name:     "Cancellation transaction with fromAddress To address",
			tx:       txtypes.NewTx(&txtypes.LegacyTx{Nonce: 0, To: &fromAddress, Value: big.NewInt(0)}),
			expected: true,
		},
		{
			name:     "Non-cancellation transaction with non-zero To address",
			tx:       txtypes.NewTx(&txtypes.LegacyTx{Nonce: 0, To: &fromAddress, Value: big.NewInt(1)}),
			expected: false,
		},
		{
			name:     "Non-cancellation transaction with non-fromAddress To address",
			tx:       txtypes.NewTx(&txtypes.LegacyTx{Nonce: 0, To: &anotherAddress, Value: big.NewInt(0)}),
			expected: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCancellationTransaction(tt.tx, fromAddress)
			if result != tt.expected {
				t.Errorf("isCancellationTransaction() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// todo refactor so common parts are not duplicated
func TestSenderAddress_LegacyTx(t *testing.T) {
	chainID := big.NewInt(1)

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("Failed to assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	expectedSender := crypto.PubkeyToAddress(*publicKeyECDSA)

	legacyTxData := &txtypes.LegacyTx{
		Nonce:    0,
		To:       &expectedSender,
		Value:    big.NewInt(100),
		Gas:      100000,
		GasPrice: big.NewInt(1),
	}

	legacyTx := txtypes.NewTx(legacyTxData)
	signer := txtypes.NewEIP155Signer(chainID)

	signedLegacyTx, err := txtypes.SignTx(legacyTx, signer, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	sender, err := SenderAddress(signedLegacyTx)
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if sender != expectedSender {
		t.Errorf("Expected sender address to be %s, got %s", expectedSender.Hex(), sender.Hex())
	}
}

func TestSenderAddress_AccessListTx(t *testing.T) {
	chainID := big.NewInt(1)

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("Failed to assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	expectedSender := crypto.PubkeyToAddress(*publicKeyECDSA)

	accessListTxData := &txtypes.AccessListTx{
		ChainID:    chainID,
		Nonce:      1,
		To:         &expectedSender,
		Value:      big.NewInt(100),
		Gas:        100000,
		GasPrice:   big.NewInt(1),
		AccessList: txtypes.AccessList{},
	}

	accessListTx := txtypes.NewTx(accessListTxData)
	signer := txtypes.NewEIP2930Signer(chainID)

	signedAccessListTx, err := txtypes.SignTx(accessListTx, signer, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	sender, err := SenderAddress(signedAccessListTx)
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if sender != expectedSender {
		t.Errorf("Expected sender address to be %s, got %s", expectedSender.Hex(), sender.Hex())
	}
}

func TestSenderAddress_DynamicFeeTx(t *testing.T) {
	chainID := big.NewInt(1)

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("Failed to assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	expectedSender := crypto.PubkeyToAddress(*publicKeyECDSA)

	dynamicFeeTxData := &txtypes.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     2,
		To:        &expectedSender,
		Value:     big.NewInt(100),
		Gas:       100000,
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(100),
	}

	dynamicFeeTx := txtypes.NewTx(dynamicFeeTxData)
	signer := txtypes.NewLondonSigner(chainID)

	signedDynamicFeeTx, err := txtypes.SignTx(dynamicFeeTx, signer, privateKey)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	sender, err := SenderAddress(signedDynamicFeeTx)
	if err != nil {
		t.Errorf("Error occurred: %v", err)
	}
	if sender != expectedSender {
		t.Errorf("Expected sender address to be %s, got %s", expectedSender.Hex(), sender.Hex())
	}
}
