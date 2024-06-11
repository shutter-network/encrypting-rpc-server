package requests

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func SendCancelTx(client *ethclient.Client, privateKey *ecdsa.PrivateKey) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get suggested gas price: %v", err)
	}

	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(11))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(10))

	// todo change to using NewTx()
	tx := types.NewTransaction(nonce, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(0), 21000, gasPrice, nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	log.Printf("Transaction sent: %s", signedTx.Hash().Hex())
}
