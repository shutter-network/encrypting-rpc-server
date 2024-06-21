package requests

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
)

func SendTx(client *rpc.Client, s string) common.Hash {
	var txHash common.Hash
	err := client.CallContext(context.Background(), &txHash, "eth_sendRawTransaction", s)
	if err != nil {
		log.Fatalf("Failed to send raw transaction: %v", err)
	}

	return txHash
}
