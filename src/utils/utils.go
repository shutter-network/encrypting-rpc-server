package utils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/joho/godotenv"
	"log"
	"math/big"
	"path/filepath"
	"runtime"
	"testing"
)

func LoadEnv() {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	rootPath := filepath.Join(basepath, "..")

	err := godotenv.Load(filepath.Join(rootPath, ".env"))
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func CheckErr(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

func SenderAddress(tx *txtypes.Transaction) (common.Address, error) {
	chainID := tx.ChainId()
	var signer txtypes.Signer

	switch tx.Type() {
	case txtypes.LegacyTxType:
		signer = txtypes.NewEIP155Signer(chainID)
	case txtypes.AccessListTxType:
		signer = txtypes.NewEIP2930Signer(chainID)
	case txtypes.DynamicFeeTxType:
		signer = txtypes.NewLondonSigner(chainID)
	default:
		return common.Address{}, fmt.Errorf("unsupported transaction type: %d", tx.Type())
	}

	sender, err := txtypes.Sender(signer, tx)
	if err != nil {
		return common.Address{}, err
	}

	return sender, nil
}

func IsCancellationTransaction(tx *txtypes.Transaction, fromAddress common.Address) bool {
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	return tx.Value().Cmp(big.NewInt(0)) == 0 && (tx.To() == nil || *tx.To() == zeroAddress || *tx.To() == fromAddress)
}
