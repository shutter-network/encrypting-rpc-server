package rpc

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"strconv"
	"strings"
)

func (s *EthService) CheckNonceUsed(key string, ctx context.Context) bool {
	keyAddress := strings.Split(key, "-")[0]
	keyNonce := strings.Split(key, "-")[1]
	nonce, err := strconv.ParseUint(keyNonce, 10, 64)
	if err == nil {
		accountNonce, err := s.Processor.Client.NonceAt(ctx, common.HexToAddress(keyAddress), nil)
		if err == nil && accountNonce > nonce {
			return true
		}
	}
	return false
}
