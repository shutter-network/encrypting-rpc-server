package stress

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"sort"
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

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}

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
	SignerForChain       types.Signer
	ChainID              *big.Int
	SubmitSign           bind.SignerFn
	SubmitPrivateKey     *ecdsa.PrivateKey
	SubmitFromAddress    common.Address
	TransactSign         bind.SignerFn
	TransactPrivateKey   *ecdsa.PrivateKey
	TransactFromAddress  common.Address
	Sequencer            sequencerBindings.Sequencer
	KeyperSetManager     shopContractBindings.KeyperSetManager
	KeyBroadcastContract shopContractBindings.KeyBroadcastContract
}

func createSetup(fundNewAccount bool) (StressSetup, error) {
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
	setup.ChainID = chainID

	signerForChain := types.LatestSignerForChainID(chainID)
	setup.SignerForChain = signerForChain

	submitKeyHex := os.Getenv("STRESS_TEST_PK")
	if len(submitKeyHex) < 64 {
		return *setup, errors.New("private key hex must be in environment variable STRESS_TEST_PK")
	}
	submitPrivateKey, err := crypto.HexToECDSA(submitKeyHex)
	if err != nil {
		return *setup, err
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

	// TODO: allow multiple transacting accounts in StressEnvironment.TransactAccounts
	transactPrivateKey, err := crypto.GenerateKey()
	transactPrivateKeyBytes := crypto.FromECDSA(transactPrivateKey)

	if err != nil {
		return *setup, err
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
		return *setup, err
	}
	defer f.Close()

	encoder := hex.NewEncoder(f)
	_, err = encoder.Write(transactPrivateKeyBytes)
	if err != nil {
		return *setup, err
	}
	_, err = f.Write([]byte(" "))
	if err != nil {
		return *setup, err
	}
	_, err = encoder.Write(transactFromAddress.Bytes())
	if err != nil {
		return *setup, err
	}
	_, err = f.Write([]byte("\n"))
	if err != nil {
		return *setup, err
	}

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
	if fundNewAccount {
		err = fund(*setup)
		if err != nil {
			return *setup, err
		}
		log.Println("Funding complete")
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
	if err != nil {
		return err
	}
	tx := types.NewTransaction(nonce, setup.TransactFromAddress, value, gasLimit, gasPrice, data)
	signedTx, err := setup.SubmitSign(setup.SubmitFromAddress, tx)
	if err != nil {
		return err
	}
	err = setup.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return err
	}
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
	WaitOnEverySubmit     bool
	// work around a bug, where decryption keys are tried in the order of identityPrefixes
	EnsureOrderedPrefixes bool
	IdentityPrefixes      []shcrypto.Block
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
		Eon:                   eon,
		EonPublicKey:          eonKey,
		WaitOnEverySubmit:     false,
		EnsureOrderedPrefixes: false,
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

func createIdentity() (shcrypto.Block, error) {
	identityPrefix, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return shcrypto.Block{}, fmt.Errorf("could not get random identityPrefix %v", err)
	}
	log.Println(hex.EncodeToString(identityPrefix[:]))

	return identityPrefix, nil
}

func encrypt(ctx context.Context, tx types.Transaction, env *StressEnvironment, setup StressSetup, count int) (*shcrypto.EncryptedMessage, shcrypto.Block, error) {

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, shcrypto.Block{}, fmt.Errorf("could not get sigma bytes %s", err)
	}

	var identityPrefix shcrypto.Block
	if count < len(env.IdentityPrefixes) {
		identityPrefix = env.IdentityPrefixes[count]
	} else {
		identityPrefix, err = createIdentity()

		if err != nil {
			return nil, identityPrefix, err
		}
	}

	identity := rpc.ComputeIdentity(identityPrefix[:], setup.SubmitFromAddress)

	var buff bytes.Buffer
	err = tx.EncodeRLP(&buff)

	if err != nil {
		return nil, identityPrefix, fmt.Errorf("failed encode RLP %v", err)
	}
	j, err := tx.MarshalJSON()
	if err != nil {
		return nil, identityPrefix, fmt.Errorf("failed to marshal json %v", err)
	}
	log.Println("tx to be encrypted", string(j[:]))
	encryptedTx := shcrypto.Encrypt(buff.Bytes(), (*shcrypto.EonPublicKey)(env.EonPublicKey), identity, sigma)
	return encryptedTx, identityPrefix, nil
}

func submitEncryptedTx(ctx context.Context, setup StressSetup, env *StressEnvironment, tx types.Transaction, count int) (*types.Transaction, error) {

	opts := env.SubmitterOpts
	log.Println("submit nonce", opts.Nonce)

	opts.Value = big.NewInt(0).Sub(tx.Cost(), tx.Value())

	encryptedTx, identityPrefix, err := encrypt(ctx, tx, env, setup, count)
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

func transact(setup StressSetup, env *StressEnvironment, count int) error {

	value := big.NewInt(1)    // in wei
	gasLimit := uint64(21000) // in units

	toAddress := setup.SubmitFromAddress
	var data []byte
	var submissions []types.Transaction
	var innerTxs []types.Transaction
	suggestedGasTipCap, err := setup.Client.SuggestGasTipCap(context.Background())
	if err != nil {
		return err
	}
	suggestedGasPrice, err := setup.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}
	feeCapAndTipCap := big.NewInt(0).Add(suggestedGasPrice, suggestedGasTipCap)

	gasFloat, _ := suggestedGasPrice.Float64()

	if env.EnsureOrderedPrefixes {

		var identityPrefixes []shcrypto.Block
		for i := 0; i < count; i++ {
			identity, err := createIdentity()
			if err != nil {
				return err
			}
			identityPrefixes = append(identityPrefixes, identity)
		}
		sort.Slice(identityPrefixes, func(i, j int) bool {
			return hex.EncodeToString(identityPrefixes[i][:]) < hex.EncodeToString(identityPrefixes[j][:])
		})
		env.IdentityPrefixes = identityPrefixes
	}

	for i := 0; i < count; i++ {

		// TODO: make gas price calculation a function StressEnvironment.GasPriceFn()

		//x := int64(gasFloat * (2. / float64(count)) * float64(i+1)) // higher delta for higher nonces
		//x := int64(gasFloat * (2. / float64(count)) * float64(count-i+1)) // lower delta for higher nonces to affect ordering
		x := int64(gasFloat * 1.5) // fixed delta
		log.Println("delta is ", x)
		delta := big.NewInt(x)
		gasFeeCap := big.NewInt(0).Add(feeCapAndTipCap, delta)
		innerNonce := env.TransactStartingNonce.Uint64() + uint64(i)
		tx := types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   setup.ChainID,
				Nonce:     innerNonce,
				GasFeeCap: gasFeeCap,
				GasTipCap: suggestedGasTipCap,
				Gas:       gasLimit,
				To:        &toAddress,
				Value:     value,
				Data:      data,
			},
		)

		signedTx, err := setup.TransactSign(setup.TransactFromAddress, tx)
		if err != nil {
			return err
		}
		innerTxs = append(innerTxs, *signedTx)
		log.Println("used nonce", signedTx.Nonce())
	}
	for i := range innerTxs {
		signedTx := innerTxs[i]
		submitNonce := big.NewInt(0).Add(env.SubmitStartingNonce, big.NewInt(int64(i)))
		env.SubmitterOpts.Nonce = submitNonce
		submitTx, err := submitEncryptedTx(context.Background(), setup, env, signedTx, i)
		if err != nil {
			return err
		}
		submissions = append(submissions, *submitTx)
		if env.WaitOnEverySubmit {
			receipt, err := bind.WaitMined(context.Background(), setup.Client, submitTx)

			if err != nil {
				return fmt.Errorf("submit not mined %s", submitTx.Hash().Hex())
			} else {
				log.Println("status", receipt.Status)
			}
		}
		log.Println("Submit tx hash", submitTx.Hash().Hex(), "Encrypted tx hash", signedTx.Hash().Hex())
	}
	for _, submitTx := range submissions {
		log.Print("waiting for submission ", submitTx.Hash().Hex())

		receipt, err := bind.WaitMined(context.Background(), setup.Client, &submitTx)
		if err != nil {
			return fmt.Errorf("error on WaitMined %s", err)
		}
		log.Println("status", receipt.Status)
		if receipt.Status != 1 {
			return fmt.Errorf("submission failed")
		}
	}
	for _, innerTx := range innerTxs {
		log.Println("waiting for inclusion ", innerTx.Hash().Hex())

		receipt, err := bind.WaitMined(context.Background(), setup.Client, &innerTx)
		if err != nil {
			return fmt.Errorf("error on WaitMined %s", err)
		}
		log.Println("status", receipt.Status)
		if receipt.Status != 1 {
			return fmt.Errorf("included tx failed")
		}
	}
	return nil
}

func ReadPks(r io.Reader) ([]*ecdsa.PrivateKey, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	var result []*ecdsa.PrivateKey
	for scanner.Scan() {
		x := scanner.Text()
		if len(x) == 64 {
			pk, err := crypto.HexToECDSA(x)
			if err != nil {
				return result, err
			}
			result = append(result, pk)
		}
	}
	return result, scanner.Err()
}

func drain(ctx context.Context, pk *ecdsa.PrivateKey, address common.Address, balance uint64, setup StressSetup) {
	gasPrice, err := setup.Client.SuggestGasPrice(ctx)
	if err != nil {
		log.Println("could not query gasPrice")
	}
	gasLimit := uint64(21000)
	remaining := balance - gasLimit*gasPrice.Uint64()
	data := make([]byte, 0)

	nonce, err := setup.Client.PendingNonceAt(ctx, address)
	if err != nil {
		log.Println("could not query nonce", err)
	}
	tx := types.NewTransaction(nonce, setup.SubmitFromAddress, big.NewInt(int64(remaining)), gasLimit, gasPrice, data)

	signature, err := crypto.Sign(setup.SignerForChain.Hash(tx).Bytes(), pk)
	if err != nil {
		log.Println("could not create signature", err)
	}
	signed, err := tx.WithSignature(setup.SignerForChain, signature)
	if err != nil {
		log.Println("could not add signature", err)
	}
	err = setup.Client.SendTransaction(ctx, signed)
	if err != nil {
		log.Println("failed to send", err)
	}
	receipt, err := bind.WaitMined(ctx, setup.Client, signed)
	if err != nil {
		log.Println("failed to wait for tx", err)
	}
	log.Println("status", receipt.Status)
}

// not really a test, but useful to collect from previously funded test accounts
func TestEmptyAccounts(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(false)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	fd, err := os.Open("pk.hex")
	if err != nil {
		log.Fatal("Could not open pk.hex")
	}
	defer fd.Close()
	pks, err := ReadPks(fd)
	if err != nil {
		log.Fatal("error when reading private keys", err)
	}
	block, err := setup.Client.BlockNumber(context.Background())
	if err != nil {
		log.Fatal("could not query block number", err)
	}
	for i := range pks {
		public := pks[i].Public()
		publicKey, ok := public.(*ecdsa.PublicKey)
		if !ok {
			log.Fatal("error casting public key to ECDSA")
		}
		address := crypto.PubkeyToAddress(*publicKey)
		balance, err := setup.Client.BalanceAt(context.Background(), address, big.NewInt(int64(block)))
		if err == nil {
			log.Println(address.Hex(), balance)
		}
		if balance.Uint64() > 0 {
			drain(context.Background(), pks[i], address, balance.Uint64(), setup)
		}
	}
}

func TestStressSingle(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	err = transact(setup, &env, 1)
	if err != nil {
		log.Printf("failure %s", err)
		t.Fail()
	}
	fmt.Println("transacted")
}

func TestStressDualWait(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	env.WaitOnEverySubmit = true

	err = transact(setup, &env, 2)
	if err != nil {
		log.Printf("failure %s", err)
		t.Fail()
	}
	fmt.Println("transacted")
}

// run with `go test -test.v -timeout 3m -run TestStressDualNoWait`; currently flaky
func TestStressDualNoWait(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}

	err = transact(setup, &env, 2)
	if err != nil {
		log.Printf("failure %s", err)
		t.Fail()
	}
	fmt.Println("transacted")
}

// send many transactions as quickly as possible, but ensure the identityPrefixes are ordered (low to high)
func TestStressManyNoWaitOrderedPrefix(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}

	env.EnsureOrderedPrefixes = true
	err = transact(setup, &env, 20)
	if err != nil {
		log.Printf("failure %s", err)
		t.Fail()
	}
	fmt.Println("transacted")
}
