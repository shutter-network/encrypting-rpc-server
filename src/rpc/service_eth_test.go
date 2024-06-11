package rpc

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	. "github.com/ovechkin-dm/mockio/mock"
	"github.com/shutter-network/encrypting-rpc-server/cache"
	"github.com/stretchr/testify/mock"
	"math/big"
	"testing"
)

// test 1 with Mockio
func TestSimple(t *testing.T) {
	SetUp(t)
	m := Mock[EthService]()
	ctx := context.Background()
	blockNumber := uint64(100)
	someAddress := common.HexToAddress("0xC0058BdcC93EaA1afd468f06A26394E2d80c8f01")

	tx1 := &txtypes.LegacyTx{
		Nonce:    0,
		To:       &someAddress,
		Value:    big.NewInt(100),
		Gas:      100000,
		GasPrice: big.NewInt(1),
	}
	tx := txtypes.NewTx(tx1)
	returnedHash := common.HexToHash("0x37a7aee34d94dbbc890b402bf49997495ef12ea5a400b7441ba946fc5e42be7b")

	When(m.SendRawTransaction(AnyContext(), AnyString())).ThenReturn(&returnedHash)

	m.cache.Data["key1"] = cache.TransactionInfo{
		SendingBlock: blockNumber,
		Tx:           tx,
	}

	m.NewBlock(ctx, blockNumber)

	fmt.Print(m.cache.Data["key1"])

	//Verify(m, AtLeastOnce()).SendRawTransaction(ctx, tx.Hash().Hex())
}

// test 2 with testify
type MockEthService struct {
	mock.Mock
	EthService
}

func (m *MockEthService) SendRawTransaction(ctx context.Context, txHash string) (*common.Hash, error) {
	args := m.Called(ctx, txHash)
	return args.Get(0).(*common.Hash), args.Error(1)
}

func TestNewBlock(t *testing.T) {
	ctx := context.Background()
	blockNumber := uint64(100)
	someAddress := common.HexToAddress("0xC0058BdcC93EaA1afd468f06A26394E2d80c8f01")

	tx1 := &txtypes.LegacyTx{
		Nonce:    0,
		To:       &someAddress,
		Value:    big.NewInt(100),
		Gas:      100000,
		GasPrice: big.NewInt(1),
	}
	tx := txtypes.NewTx(tx1)

	mockService := new(MockEthService)

	mockService.EthService.cache = &cache.Cache{
		Data: map[string]cache.TransactionInfo{
			"key1": {
				SendingBlock: blockNumber,
				Tx:           tx,
			},
			"key2": {
				SendingBlock: blockNumber,
				Tx:           nil,
			},
		},
		DelayFactor: 10,
	}

	returnedHash := common.HexToHash("0x37a7aee34d94dbbc890b402bf49997495ef12ea5a400b7441ba946fc5e42be7b")

	mockService.On("SendRawTransaction", ctx, tx.Hash().Hex()).Return(&returnedHash, nil)

	mockService.NewBlock(ctx, blockNumber)

	mockService.AssertExpectations(t)
	//require.NotContains(t, mockService.cache.Data, "key2", "Expected key2 to be deleted from cache")
	//require.Equal(t, blockNumber+mockService.cache.DelayFactor, mockService.cache.Data["key1"].SendingBlock, "Expected SendingBlock to be updated")
}
