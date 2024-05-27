package stress

import (
	"context"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shutter-network/encrypting-rpc-server/rpc"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	shopContractBindings "github.com/shutter-network/shop-contracts/bindings"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

// the ethereum address of the key broadcast contract
const KEY_BROADCAST_CONTRACT_ADDRESS = "0x1FD85EfeC5FC18f2f688f82489468222dfC36d6D"

// the ethereum address of the sequencer contract
const SEQUENCER_CONTRACT_ADDRESS = "0xd073BD5A717Dce1832890f2Fdd9F4fBC4555e41A"

// the ethereum address of the keyper set manager contract
const KEYPER_SET_MANAGER_CONTRACT_ADDRESS = "0x7Fbc29C682f59f809583bFEE0fc50F1e4eb77774"

const KeyperSetChangeLookAhead = 2

func encrypt(ctx context.Context, client *ethclient.Client, tx types.Transaction, pk *ecdsa.PrivateKey) (*shcrypto.EncryptedMessage, uint64, shcrypto.Block) {
	blockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatal("could not query blockNumber", err)
	}

	keyperSetManagerContract, err := shopContractBindings.NewKeyperSetManager(common.HexToAddress(KEYPER_SET_MANAGER_CONTRACT_ADDRESS), client)
	if err != nil {
		log.Fatal("can not get KeyperSetManager", err)
	}
	eon, err := keyperSetManagerContract.GetKeyperSetIndexByBlock(nil, blockNumber+uint64(KeyperSetChangeLookAhead))
	if err != nil {
		log.Fatal("could not get eon", err)
	}

	keyBroadcastContract, err := shopContractBindings.NewKeyBroadcastContract(common.HexToAddress(KEY_BROADCAST_CONTRACT_ADDRESS), client)
	if err != nil {
		log.Fatal("can not get KeyBrodcastContract", err)
	}

	eonKeyBytes, err := keyBroadcastContract.GetEonKey(nil, eon)
	if err != nil {
		log.Fatal("could not get eonKeyBytes", err)
	}

	eonKey := &shcrypto.EonPublicKey{}
	if err := eonKey.Unmarshal(eonKeyBytes); err != nil {
		log.Fatal("could not unmarshal eonKeyBytes", err)
	}

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		log.Fatal("could not get sigma bytes", err)
	}

	chainId, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal("could not get ChainId", err)
	}

	newSigner, err := bind.NewKeyedTransactorWithChainID(pk, chainId)
	if err != nil {
		log.Fatal("could not get signer", err)
	}

	identityPrefix, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		log.Fatal("could not get random identityPrefix", err)
	}
	identity := rpc.ComputeIdentity(identityPrefix[:], newSigner.From)
	b, err := tx.MarshalJSON()
	encryptedTx := shcrypto.Encrypt(b, eonKey, identity, sigma)
	return encryptedTx, eon, identityPrefix
}

func submitEncryptedTx(ctx context.Context, tx types.Transaction, from *common.Address, pk *ecdsa.PrivateKey, client *ethclient.Client) {
	chainId, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal("failed retrieve chainId", err)
	}
	newSigner, err := bind.NewKeyedTransactorWithChainID(pk, chainId)
	if err != nil {
		log.Fatal("failed to create signer", err)
	}

	opts := bind.TransactOpts{
		From:   *from,
		Signer: newSigner.Signer,
	}

	opts.Value = big.NewInt(0).Sub(tx.Cost(), tx.Value())

	sequencerContract, err := sequencerBindings.NewSequencer(common.HexToAddress(SEQUENCER_CONTRACT_ADDRESS), client)

	encryptedTx, eon, identityPrefix := encrypt(ctx, client, tx, pk)

	// TODO: set opts.Nonce !
	submitTx, err := sequencerContract.SubmitEncryptedTransaction(&opts, eon, identityPrefix, encryptedTx.Marshal(), new(big.Int).SetUint64(tx.Gas()))
	if err != nil {
		log.Fatal("Could not submit", err)
	}
	log.Println("Sent tx with hash", tx.Hash().Hex(), "Encrypted tx hash", submitTx.Hash().Hex())
	_, err = bind.WaitMined(ctx, client, submitTx)
	if err != nil {
		log.Fatal("error on WaitMined", err)
	}

}

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
	log.Println("Current nonce is", nonce)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1)    // in wei
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0xF1fc0e5B6C5E42639d27ab4f2860e964de159bB4")
	var data []byte
	tx := types.NewTransaction(nonce+1, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	submitEncryptedTx(context.Background(), *signedTx, &fromAddress, privateKey, client)

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

/* TODO:

- transact batches
- plan nonces: for i in len(batch)
  - submit nonces = latest nonce + i
  - encrypted nonces = latest nonce + len(batch) + i

- collect tx hashes, wait for mined only after batch is submitted

*/
