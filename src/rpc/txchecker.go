package rpc

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
)

func CalculateIntrinsicGas(tx *types.Transaction) (uint64, error) {
	isContractCreation := tx.To() == nil

	intrinsicGas, err := core.IntrinsicGas(tx.Data(), tx.AccessList(), isContractCreation, true, true, true)
	if err != nil {
		return 0, err
	}

	return intrinsicGas, nil
}
