package stress

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const SEQUENCER_CONTRACT_ADDRESS = "0xd073BD5A717Dce1832890f2Fdd9F4fBC4555e41A"

func transact() {
	client, err := ethclient.Dial("https://rpc.chiado.gnosis.gateway.fm")
	if err != nil {
		log.Fatal(err)
	}

	keyHex := os.Getenv("STRESS_TEST_PK")
	if len(keyHex) < 64 {
		log.Fatal("private key hex must be in environment variable STRESS_TEST_PK")
	}
	privateKey, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1)    // in wei (1 eth)
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0xF1fc0e5B6C5E42639d27ab4f2860e964de159bB4")
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
}

func TestStress(t *testing.T) {
	fmt.Println("Hello, World!")
	transact()
	fmt.Println("transacted")
}
