package rpc

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestCalculateIntrinsicGas_SimpleEtherTransfer(t *testing.T) {
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{},
		Value:    big.NewInt(1000000000000000000),
		Gas:      21000,
		GasPrice: big.NewInt(1000000000),
		Data:     nil,
	})
	expectedGas := uint64(21000)

	gas, err := CalculateIntrinsicGas(tx)

	assert.NoError(t, err)
	assert.Equal(t, expectedGas, gas)
}

func TestCalculateIntrinsicGas_ContractCreation(t *testing.T) {
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    0,
		To:       nil,
		Value:    big.NewInt(0),
		Gas:      53000,
		GasPrice: big.NewInt(1000000000),
		Data:     []byte{0x60, 0x60, 0x60, 0x40, 0x52, 0x60, 0x40, 0xf3},
	})
	expectedGas := uint64(53130) // 5300 + 2 for contract creation + data gas (8 * 16)

	gas, err := CalculateIntrinsicGas(tx)

	assert.NoError(t, err)
	assert.Equal(t, expectedGas, gas)
}

func TestCalculateIntrinsicGas_TransactionWithData(t *testing.T) {
	toAddress := common.HexToAddress("0x7eFf8b8A921Bd6E342042F05d0d5C0424A1f7a75")

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    0,
		To:       &toAddress,
		Value:    big.NewInt(1000000000000000000),
		Gas:      21000,
		GasPrice: big.NewInt(1000000000),
		Data:     []byte{0x12, 0x34, 0x56, 0x00},
	})
	expectedGas := uint64(21052) // 21000 base gas + (3 * 16) + (1 * 4)

	gas, err := CalculateIntrinsicGas(tx)

	assert.NoError(t, err)
	assert.Equal(t, expectedGas, gas)
}
