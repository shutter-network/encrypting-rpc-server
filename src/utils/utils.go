package utils

import (
	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func setUpLogger() zerolog.Logger {
	logger := zerolog.New(os.Stdout).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Logger()
	return logger
}

var Logger = setUpLogger()

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
	signer := txtypes.NewLondonSigner(tx.ChainId())
	fromAddress, err := signer.Sender(tx)
	if err != nil {
		return common.Address{}, err
	}

	return fromAddress, nil
}

func IsCancellationTransaction(tx *txtypes.Transaction, fromAddress common.Address) bool {
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	return tx.Value().Cmp(big.NewInt(0)) == 0 && (tx.To() == nil || *tx.To() == zeroAddress || *tx.To() == fromAddress)
}
