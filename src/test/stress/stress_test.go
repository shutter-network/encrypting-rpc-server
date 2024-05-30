package stress

import (
	"context"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
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

// contains all the setup required to interact with the chain
type StressSetup struct {
	Client               *ethclient.Client
	Signer               types.EIP155Signer
	Sign                 bind.SignerFn
	PrivateKey           *ecdsa.PrivateKey
	FromAddress          common.Address
	Sequencer            sequencerBindings.Sequencer
	KeyperSetManager     shopContractBindings.KeyperSetManager
	KeyBroadcastContract shopContractBindings.KeyBroadcastContract
}

func createSetup() (StressSetup, error) {
	setup := new(StressSetup)
	client, err := ethclient.Dial("https://rpc.chiado.gnosis.gateway.fm")
	if err != nil {
		return *setup, fmt.Errorf("could not create client %v", err)
	}

	keyHex := os.Getenv("STRESS_TEST_PK")
	if len(keyHex) < 64 {
		return *setup, errors.New("private key hex must be in environment variable STRESS_TEST_PK")
	}
	privateKey, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return *setup, errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return *setup, fmt.Errorf("could not query chainId %v", err)
	}

	keyperSetManagerContract, err := shopContractBindings.NewKeyperSetManager(common.HexToAddress(KEYPER_SET_MANAGER_CONTRACT_ADDRESS), client)
	if err != nil {
		return *setup, fmt.Errorf("can not get KeyperSetManager %v", err)
	}
	setup.KeyperSetManager = *keyperSetManagerContract

	keyBroadcastContract, err := shopContractBindings.NewKeyBroadcastContract(common.HexToAddress(KEY_BROADCAST_CONTRACT_ADDRESS), client)
	if err != nil {
		return *setup, fmt.Errorf("can not get KeyBrodcastContract %v", err)
	}

	setup.KeyBroadcastContract = *keyBroadcastContract

	sequencerContract, err := sequencerBindings.NewSequencer(common.HexToAddress(SEQUENCER_CONTRACT_ADDRESS), client)
	if err != nil {
		return *setup, fmt.Errorf("can not get SequencerContract %v", err)
	}

	setup.Sequencer = *sequencerContract

	signerForChain := types.LatestSignerForChainID(chainID)
	setup.Client = client
	setup.Sign = func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != fromAddress {
			return nil, errors.New("Not Authorized")
		}
		signature, err := crypto.Sign(signerForChain.Hash(tx).Bytes(), privateKey)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(signerForChain, signature)
	}
	setup.PrivateKey = privateKey
	setup.FromAddress = fromAddress

	return *setup, nil
}

// contains the context for the current stress test to create transactions
type StressEnvironment struct {
	Opts          bind.TransactOpts
	StartingNonce *big.Int
	Eon           uint64
	EonPublicKey  *shcrypto.EonPublicKey
}

func createStressEnvironment(ctx context.Context, setup StressSetup) (StressEnvironment, error) {
	eon, eonKey, err := getEonKey(ctx, setup)

	environment := StressEnvironment{
		Opts: bind.TransactOpts{
			From:   setup.FromAddress,
			Signer: setup.Sign,
		},
		Eon:          eon,
		EonPublicKey: eonKey,
	}
	if err != nil {
		return environment, fmt.Errorf("could not get eonKey %v", err)
	}
	nonce, err := setup.Client.PendingNonceAt(context.Background(), setup.FromAddress)
	log.Println("Current nonce is", nonce)
	if err != nil {
		return environment, fmt.Errorf("could not query starting nonce %v", err)
	}
	environment.StartingNonce = big.NewInt(int64(nonce))
	log.Println("eon is ", eon)
	return environment, nil
}

func getEonKey(ctx context.Context, setup StressSetup) (uint64, *shcrypto.EonPublicKey, error) {
	blockNumber, err := setup.Client.BlockNumber(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("could not query blockNumber %v", err)
	}

	eon, err := setup.KeyperSetManager.GetKeyperSetIndexByBlock(nil, blockNumber+uint64(KeyperSetChangeLookAhead))
	if err != nil {
		return 0, nil, fmt.Errorf("could not get eon %v", err)
	}

	eonKeyBytes, err := setup.KeyBroadcastContract.GetEonKey(nil, eon)
	if err != nil {
		return 0, nil, fmt.Errorf("could not get eonKeyBytes %v", err)
	}

	eonKey := &shcrypto.EonPublicKey{}
	if err := eonKey.Unmarshal(eonKeyBytes); err != nil {
		return 0, nil, fmt.Errorf("could not unmarshal eonKeyBytes %v", err)
	}
	return eon, eonKey, nil

}

func encrypt(ctx context.Context, tx types.Transaction, env StressEnvironment, setup StressSetup) (*shcrypto.EncryptedMessage, shcrypto.Block, error) {

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, shcrypto.Block{}, fmt.Errorf("could not get sigma bytes %s", err)
	}

	identityPrefix, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, shcrypto.Block{}, fmt.Errorf("could not get random identityPrefix %v", err)
	}
	identity := rpc.ComputeIdentity(identityPrefix[:], setup.FromAddress)
	identityMarshal := identity.Marshal()

	log.Print("creating Identity ", hex.EncodeToString(identityMarshal))
	b, err := tx.MarshalJSON()
	if err != nil {
		return nil, identityPrefix, fmt.Errorf("failed to marshal tx %v", err)
	}
	encryptedTx := shcrypto.Encrypt(b, (*shcrypto.EonPublicKey)(env.EonPublicKey), identity, sigma)
	return encryptedTx, identityPrefix, nil
}

func submitEncryptedTx(ctx context.Context, setup StressSetup, env StressEnvironment, tx types.Transaction) (*types.Transaction, error) {

	opts := env.Opts

	opts.Value = big.NewInt(0).Sub(tx.Cost(), tx.Value())

	encryptedTx, identityPrefix, err := encrypt(ctx, tx, env, setup)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt %v", err)
	}

	submitTx, err := setup.Sequencer.SubmitEncryptedTransaction(&opts, env.Eon, identityPrefix, encryptedTx.Marshal(), new(big.Int).SetUint64(tx.Gas()))
	if err != nil {
		return nil, fmt.Errorf("Could not submit %s", err)
	}
	identityHex, _ := identityPrefix.MarshalText()
	log.Println("submitted identityPrefix ", hex.EncodeToString(identityHex))
	return submitTx, nil

}

func transact(setup StressSetup) {

	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1)    // in wei
	gasLimit := uint64(21000) // in units
	gasPrice, err := setup.Client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress("0xF1fc0e5B6C5E42639d27ab4f2860e964de159bB4")
	var data []byte
	var count = 2
	var submissions []types.Transaction
	var innerTxs []types.Transaction
	for i := 0; i < count; i++ {

		innerNonce := env.StartingNonce.Uint64() + uint64(count) + uint64(i)
		tx := types.NewTransaction(innerNonce, toAddress, value, gasLimit, gasPrice, data)

		signedTx, err := setup.Sign(setup.FromAddress, tx)
		if err != nil {
			log.Fatal(err)
		}
		innerTxs = append(innerTxs, *signedTx)
		env.Opts.Nonce = big.NewInt(0).Add(env.StartingNonce, big.NewInt(int64(i)))
		submitTx, err := submitEncryptedTx(context.Background(), setup, env, *signedTx)
		if err != nil {
			log.Fatal(err)
		}
		submissions = append(submissions, *submitTx)
		log.Println("Sent tx with hash", submitTx.Hash().Hex(), "Encrypted tx hash", signedTx.Hash().Hex(), "Original tx hash", tx.Hash().Hex())
	}
	for _, submitTx := range submissions {
		log.Println("waiting for submission ", submitTx.Hash().Hex())

		_, err = bind.WaitMined(context.Background(), setup.Client, &submitTx)
		if err != nil {
			log.Fatal("error on WaitMined", err)
		}
	}
	for _, innerTx := range innerTxs {
		log.Println("waiting for inclusion ", innerTx.Hash().Hex())

		_, err = bind.WaitMined(context.Background(), setup.Client, &innerTx)
		if err != nil {
			log.Fatal("error on WaitMined", err)
		}
	}
}

func TestStress(t *testing.T) {
	fmt.Println("Hello, World!")
	setup, err := createSetup()
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	transact(setup)
	fmt.Println("transacted")
}
