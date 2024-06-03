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
	SubmitSigner         types.EIP155Signer
	SubmitSign           bind.SignerFn
	SubmitPrivateKey     *ecdsa.PrivateKey
	SubmitFromAddress    common.Address
	TransactSigner       types.EIP155Signer
	TransactSign         bind.SignerFn
	TransactPrivateKey   *ecdsa.PrivateKey
	TransactFromAddress  common.Address
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

	setup.Client = client

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return *setup, fmt.Errorf("could not query chainId %v", err)
	}

	signerForChain := types.LatestSignerForChainID(chainID)

	submitKeyHex := os.Getenv("STRESS_TEST_PK")
	if len(submitKeyHex) < 64 {
		return *setup, errors.New("private key hex must be in environment variable STRESS_TEST_PK")
	}
	submitPrivateKey, err := crypto.HexToECDSA(submitKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	submitPublicKey := submitPrivateKey.Public()
	submitPublicKeyECDSA, ok := submitPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return *setup, errors.New("error casting public key to ECDSA")
	}

	submitFromAddress := crypto.PubkeyToAddress(*submitPublicKeyECDSA)

	setup.SubmitSign = func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != submitFromAddress {
			return nil, errors.New("Not Authorized")
		}
		signature, err := crypto.Sign(signerForChain.Hash(tx).Bytes(), submitPrivateKey)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(signerForChain, signature)
	}
	setup.SubmitPrivateKey = submitPrivateKey
	setup.SubmitFromAddress = submitFromAddress

	transactPrivateKey, err := crypto.GenerateKey()
	transactPrivateKeyBytes := crypto.FromECDSA(transactPrivateKey)

	if err != nil {
		log.Fatal(err)
	}

	transactPublicKey := transactPrivateKey.Public()
	transactPublicKeyECDSA, ok := transactPublicKey.(*ecdsa.PublicKey)
	if !ok {
		return *setup, errors.New("error casting public key to ECDSA")
	}

	transactFromAddress := crypto.PubkeyToAddress(*transactPublicKeyECDSA)

	// we're going to store the privatekey of the secondary address in a file 'pk.hex'
	// this will allow us to recover funds, in case the clean up step fails
	f, err := os.OpenFile("pk.hex", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	encoder := hex.NewEncoder(f)
	encoder.Write(transactPrivateKeyBytes)
	f.Write([]byte(" "))
	encoder.Write(transactFromAddress.Bytes())
	f.Write([]byte("\n"))

	setup.TransactSign = func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != transactFromAddress {
			return nil, errors.New("Not Authorized")
		}
		signature, err := crypto.Sign(signerForChain.Hash(tx).Bytes(), transactPrivateKey)
		if err != nil {
			return nil, err
		}
		return tx.WithSignature(signerForChain, signature)
	}
	setup.TransactPrivateKey = transactPrivateKey
	setup.TransactFromAddress = transactFromAddress

	err = fund(*setup)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Funding complete")

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

	return *setup, nil
}

func fund(setup StressSetup) error {
	value := big.NewInt(100000000000000000) // 0.1 ETH in wei
	gasLimit := uint64(21000)
	gasPrice, err := setup.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	var data []byte
	nonce, err := setup.Client.PendingNonceAt(context.Background(), setup.SubmitFromAddress)
	tx := types.NewTransaction(nonce, setup.TransactFromAddress, value, gasLimit, gasPrice, data)
	signedTx, err := setup.SubmitSign(setup.SubmitFromAddress, tx)
	setup.Client.SendTransaction(context.Background(), signedTx)
	log.Println("sent funding tx", signedTx.Hash().Hex(), "to", setup.TransactFromAddress)
	_, err = bind.WaitMined(context.Background(), setup.Client, signedTx)
	return err
}

// contains the context for the current stress test to create transactions
type StressEnvironment struct {
	TransacterOpts        bind.TransactOpts
	TransactStartingNonce *big.Int
	SubmitterOpts         bind.TransactOpts
	SubmitStartingNonce   *big.Int
	Eon                   uint64
	EonPublicKey          *shcrypto.EonPublicKey
}

func createStressEnvironment(ctx context.Context, setup StressSetup) (StressEnvironment, error) {
	eon, eonKey, err := getEonKey(ctx, setup)

	environment := StressEnvironment{
		TransacterOpts: bind.TransactOpts{
			From:   setup.TransactFromAddress,
			Signer: setup.TransactSign,
		},
		SubmitterOpts: bind.TransactOpts{
			From:   setup.SubmitFromAddress,
			Signer: setup.SubmitSign,
		},
		Eon:          eon,
		EonPublicKey: eonKey,
	}
	if err != nil {
		return environment, fmt.Errorf("could not get eonKey %v", err)
	}
	submitterNonce, err := setup.Client.PendingNonceAt(context.Background(), setup.SubmitFromAddress)
	log.Println("Current submitter nonce is", submitterNonce)
	if err != nil {
		return environment, fmt.Errorf("could not query starting nonce %v", err)
	}
	environment.SubmitStartingNonce = big.NewInt(int64(submitterNonce))

	transactNonce, err := setup.Client.PendingNonceAt(context.Background(), setup.TransactFromAddress)
	log.Println("Current transacter nonce is", transactNonce)
	if err != nil {
		return environment, fmt.Errorf("could not query starting nonce %v", err)
	}
	environment.TransactStartingNonce = big.NewInt(int64(transactNonce))

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
	identity := rpc.ComputeIdentity(identityPrefix[:], setup.SubmitFromAddress)
	identityMarshal := identity.Marshal()

	log.Println("creating Identity ", hex.EncodeToString(identityMarshal))
	log.Println("nonce before encryption", tx.Nonce())
	b, err := tx.MarshalJSON()
	if err != nil {
		return nil, identityPrefix, fmt.Errorf("failed to marshal tx %v", err)
	}

	log.Println("json tx", string(b[:]))

	encryptedTx := shcrypto.Encrypt(b, (*shcrypto.EonPublicKey)(env.EonPublicKey), identity, sigma)
	return encryptedTx, identityPrefix, nil
}

func submitEncryptedTx(ctx context.Context, setup StressSetup, env StressEnvironment, tx types.Transaction) (*types.Transaction, error) {

	opts := env.SubmitterOpts

	opts.Value = big.NewInt(0).Sub(tx.Cost(), tx.Value())

	encryptedTx, identityPrefix, err := encrypt(ctx, tx, env, setup)
	if err != nil {
		return nil, fmt.Errorf("could not encrypt %v", err)
	}

	submitTx, err := setup.Sequencer.SubmitEncryptedTransaction(&opts, env.Eon, identityPrefix, encryptedTx.Marshal(), new(big.Int).SetUint64(tx.Gas()))
	if err != nil {
		return nil, fmt.Errorf("Could not submit %s", err)
	}
	log.Println("submitted identityPrefix ", hex.EncodeToString(identityPrefix[:]))
	return submitTx, nil

}

func transact(setup StressSetup, count int) {

	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(1)    // in wei
	gasLimit := uint64(21000) // in units

	toAddress := setup.SubmitFromAddress
	var data []byte
	var submissions []types.Transaction
	var innerTxs []types.Transaction
	suggestedGasTipCap, err := setup.Client.SuggestGasTipCap(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	suggestedGasPrice, err := setup.Client.SuggestGasPrice(context.Background())

	asFloat, _ := suggestedGasPrice.Float64()

	x := int64(asFloat * .15)
	delta := big.NewInt(x)
	feeCapAndTipCap := big.NewInt(0).Add(suggestedGasPrice, suggestedGasTipCap)
	gasFeeCap := big.NewInt(0).Add(feeCapAndTipCap, delta)
	for i := 0; i < count; i++ {

		innerNonce := env.TransactStartingNonce.Uint64() + uint64(i)
		tx := types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   setup.SubmitSigner.ChainID(),
				Nonce:     innerNonce,
				GasFeeCap: gasFeeCap,
				GasTipCap: suggestedGasTipCap,
				Gas:       gasLimit,
				To:        &toAddress,
				Value:     value,
				Data:      data,
			},

		//	innerNonce, toAddress, value, gasLimit, gasPrice, data
		)

		signedTx, err := setup.TransactSign(setup.TransactFromAddress, tx)
		if err != nil {
			log.Fatal(err)
		}
		innerTxs = append(innerTxs, *signedTx)
		env.TransacterOpts.Nonce = big.NewInt(0).Add(env.TransactStartingNonce, big.NewInt(int64(i)))
		log.Println("new nonce is", env.TransacterOpts.Nonce, "used nonce is", signedTx.Nonce())
		submitTx, err := submitEncryptedTx(context.Background(), setup, env, *signedTx)
		if err != nil {
			log.Fatal(err)
		}
		submissions = append(submissions, *submitTx)
		log.Println("Submit tx hash", submitTx.Hash().Hex(), "Encrypted tx hash", signedTx.Hash().Hex())
	}
	for _, submitTx := range submissions {
		log.Print("waiting for submission ", submitTx.Hash().Hex())

		receipt, err := bind.WaitMined(context.Background(), setup.Client, &submitTx)
		if err != nil {
			log.Fatal("error on WaitMined", err)
		}
		log.Println("status", receipt.Status)
		if receipt.Status != 1 {
			log.Fatal("submission failed")
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
	transact(setup, 1)
	fmt.Println("transacted")
}
