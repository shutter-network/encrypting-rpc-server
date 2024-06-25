package requests

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
)

func FetchNewBlocks(webSocketURL, httpListenAddress string) {
	client, err := ethclient.Dial(webSocketURL)

	if err != nil {
		log.Fatal(err)
	}

	headers := make(chan *types.Header)

	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatalf("Failed to subscribe to new head: %v", err) // todo potential point of failure
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:

			blockHash := header.Hash()
			fmt.Println("Block Hash:", blockHash)
			fmt.Println("Block Hash to Hex:", blockHash.Hex())

			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				if errors.Is(err, ethereum.NotFound) {
					log.Fatalf("Block not found: %v", err)
				} else {
					log.Fatalf("Failed to retrieve block: %v", err)
				}
			}

			fmt.Printf("Block number: %d\n", block.Number().Uint64())

			blockNumber := block.Number().Uint64()
			// todo reorgs? reintroduce the status field?
			SendNewBlock(blockNumber, httpListenAddress)
		}
	}
}
