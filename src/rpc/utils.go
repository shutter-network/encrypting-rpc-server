package rpc

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"os"

	"github.com/rs/zerolog"
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

func IsCancellationTransaction(tx *txtypes.Transaction, fromAddress common.Address) bool {
	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")
	return tx.Value().Cmp(big.NewInt(0)) == 0 && (tx.To() == nil || *tx.To() == zeroAddress || *tx.To() == fromAddress)
}

func SenderAddress(tx *txtypes.Transaction, chainID *big.Int) (common.Address, error) {
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
