package utils

import (
	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"
	"math/big"
	"os"
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
