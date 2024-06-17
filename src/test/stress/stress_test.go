package stress

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	cryptorand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
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
const KeyBroadcastContractAddress = "0x1FD85EfeC5FC18f2f688f82489468222dfC36d6D"

// the ethereum address of the sequencer contract
const SequencerContractAddress = "0xd073BD5A717Dce1832890f2Fdd9F4fBC4555e41A"

// the ethereum address of the keyper set manager contract
const KeyperSetManagerContractAddress = "0x7Fbc29C682f59f809583bFEE0fc50F1e4eb77774"

const RpcUrl = "https://rpc.chiado.gnosis.gateway.fm"

const KeyperSetChangeLookAhead = 2

func createSetup(fundNewAccount bool) (StressSetup, error) {
	setup := new(StressSetup)
	client, err := ethclient.Dial(RpcUrl)
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
	keyperSetManagerContract, err := shopContractBindings.NewKeyperSetManager(common.HexToAddress(KeyperSetManagerContractAddress), client)
	if err != nil {
		return *setup, fmt.Errorf("can not get KeyperSetManager %v", err)
	}
	setup.KeyperSetManager = *keyperSetManagerContract

	keyBroadcastContract, err := shopContractBindings.NewKeyBroadcastContract(common.HexToAddress(KeyBroadcastContractAddress), client)
	if err != nil {
		return *setup, fmt.Errorf("can not get KeyBrodcastContract %v", err)
	}

	setup.KeyBroadcastContract = *keyBroadcastContract

	sequencerContract, err := sequencerBindings.NewSequencer(common.HexToAddress(SequencerContractAddress), client)
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
	nonce, err := setup.Client.NonceAt(context.Background(), setup.SubmitFromAddress, nil)
	if err != nil {
		return err
	}
	log.Println("HeadNonce", nonce)
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

func defaultGasPriceFn(suggestedGasTipCap *big.Int, suggestedGasPrice *big.Int, i int, count int) (GasFeeCap, GasTipCap) {
	feeCapAndTipCap := big.NewInt(0).Add(suggestedGasPrice, suggestedGasTipCap)

	gasFloat, _ := suggestedGasPrice.Float64()
	x := int64(gasFloat * 1.5) // fixed delta
	log.Println("delta is ", x)
	delta := big.NewInt(x)
	gasFeeCap := big.NewInt(0).Add(feeCapAndTipCap, delta)
	return gasFeeCap, suggestedGasTipCap
}

func increasingGasPriceFn(suggestedGasTipCap *big.Int, suggestedGasPrice *big.Int, i int, count int) (GasFeeCap, GasTipCap) {
	feeCapAndTipCap := big.NewInt(0).Add(suggestedGasPrice, suggestedGasTipCap)

	gasFloat, _ := suggestedGasPrice.Float64()
	x := int64(gasFloat * (2. / float64(count)) * float64(i+1)) // higher delta for higher nonces
	log.Println("delta is ", x)
	delta := big.NewInt(x)
	gasFeeCap := big.NewInt(0).Add(feeCapAndTipCap, delta)
	return gasFeeCap, suggestedGasTipCap
}

func decreasingGasPriceFn(suggestedGasTipCap *big.Int, suggestedGasPrice *big.Int, i int, count int) (GasFeeCap, GasTipCap) {
	feeCapAndTipCap := big.NewInt(0).Add(suggestedGasPrice, suggestedGasTipCap)

	gasFloat, _ := suggestedGasPrice.Float64()
	x := int64(gasFloat * (2. / float64(count)) * float64(count-i)) // lower delta for higher nonces to test cut off
	log.Println("delta is ", x)
	delta := big.NewInt(x)
	gasFeeCap := big.NewInt(0).Add(feeCapAndTipCap, delta)
	return gasFeeCap, suggestedGasTipCap
}

func defaultGasLimitFn(value *big.Int, data []byte, toAddress *common.Address, i int, count int) uint64 {
	return uint64(21000)
}

func createStressEnvironment(ctx context.Context, setup StressSetup) (StressEnvironment, error) {
	eon, eonKey, err := getEonKey(ctx, setup)

	environment := StressEnvironment{
		TransacterOpts: bind.TransactOpts{
			From:   setup.TransactFromAddress,
			Signer: setup.TransactSign,
		},
		TransactGasPriceFn:   defaultGasPriceFn,
		TransactGasLimitFn:   defaultGasLimitFn,
		InclusionWaitTimeout: time.Duration(time.Minute * 2),
		InclusionConstraints: func(inclusions []*types.Receipt) error { return nil },
		SubmitterOpts: bind.TransactOpts{
			From:   setup.SubmitFromAddress,
			Signer: setup.SubmitSign,
		},
		SubmissionWaitTimeout: time.Duration(time.Second * 30),
		Eon:                   eon,
		EonPublicKey:          eonKey,
		WaitOnEverySubmit:     false,
		RandomIdentitySuffix:  false,
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
	return identityPrefix, nil
}

func encrypt(ctx context.Context, tx types.Transaction, env *StressEnvironment, submitter common.Address, i int) (*shcrypto.EncryptedMessage, shcrypto.Block, error) {

	sigma, err := shcrypto.RandomSigma(cryptorand.Reader)
	if err != nil {
		return nil, shcrypto.Block{}, fmt.Errorf("could not get sigma bytes %s", err)
	}

	var identityPrefix shcrypto.Block
	if i < len(env.IdentityPrefixes) {
		identityPrefix = env.IdentityPrefixes[i]
	} else {
		identityPrefix, err = createIdentity()

		if err != nil {
			return nil, identityPrefix, err
		}
	}
	if env.RandomIdentitySuffix {
		submitter, err = createRandomAddress()
		if err != nil {
			return nil, identityPrefix, err
		}
	}

	identity := rpc.ComputeIdentity(identityPrefix[:], submitter)

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

func submitEncryptedTx(ctx context.Context, setup StressSetup, env *StressEnvironment, tx types.Transaction, i int) (*types.Transaction, error) {

	opts := env.SubmitterOpts
	log.Println("submit nonce", opts.Nonce)

	opts.Value = big.NewInt(0).Sub(tx.Cost(), tx.Value())

	encryptedTx, identityPrefix, err := encrypt(ctx, tx, env, setup.SubmitFromAddress, i)
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

	value := big.NewInt(1) // in wei

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

	identityPrefixes := env.IdentityPrefixes
	for i := len(identityPrefixes); i < count; i++ {
		identity, err := createIdentity()
		if err != nil {
			return err
		}
		identityPrefixes = append(identityPrefixes, identity)
	}

	env.IdentityPrefixes = identityPrefixes

	for i := 0; i < count; i++ {
		gasFeeCap, suggestedGasTipCap := env.TransactGasPriceFn(suggestedGasTipCap, suggestedGasPrice, i, count)
		gasLimit := env.TransactGasLimitFn(value, data, &toAddress, i, count)
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
			_, err = waitForTx(*submitTx, "submission", env.SubmissionWaitTimeout, setup)
			if err != nil {
				return err
			}
		}
		log.Println("Submit tx hash", submitTx.Hash().Hex(), "Encrypted tx hash", signedTx.Hash().Hex())
	}
	for _, submitTx := range submissions {
		_, err = waitForTx(submitTx, "submission", env.SubmissionWaitTimeout, setup)
		if err != nil {
			return err
		}
	}
	var receipts []*types.Receipt
	for _, innerTx := range innerTxs {
		receipt, err := waitForTx(innerTx, "inclusion", env.InclusionWaitTimeout, setup)
		if err != nil {
			return err
		}
		receipts = append(receipts, receipt)
	}
	err = env.InclusionConstraints(receipts)
	if err != nil {
		return err
	}
	err = countAndLog(receipts)
	return err
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
}

// run with `go test -test.v -run TestStressDualNoWait`
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
}

// send two transactions in the same block by the same sender with the same identityPrefix
func TestStressDualDuplicatePrefix(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	prefix, err := createIdentity()
	var prefixes []shcrypto.Block
	prefixes = append(prefixes, prefix)
	prefixes = append(prefixes, prefix)
	env.IdentityPrefixes = prefixes

	err = transact(setup, &env, 2)
	if err != nil {
		log.Printf("failure %s", err)
		t.Fail()
	}
}

// send many transactions as quickly as possible.
func TestStressManyNoWait(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}

	err = transact(setup, &env, 47)
	if err != nil {
		log.Printf("failure %s", err)
		t.Fail()
	}
}

func TestStressExceedEncryptedGasLimit(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}

	env.TransactGasLimitFn = func(value *big.Int, data []byte, toAddress *common.Address, i, count int) uint64 {
		// last consumes over the limit
		if count-i == 1 {
			return uint64(1_000_000 - (i * 21_000) + 1)
		}
		return uint64(21000)
	}
	env.InclusionConstraints = func(receipts []*types.Receipt) error {
		sort.Slice(receipts, func(a, b int) bool {
			return receipts[a].BlockNumber.Uint64() < receipts[b].BlockNumber.Uint64()
		})
		if receipts[0].BlockNumber.Uint64() == receipts[len(receipts)-1].BlockNumber.Uint64() {
			return fmt.Errorf("tx must not be all in the same block")
		}
		return nil
	}
	err = transact(setup, &env, 2)
	if err != nil {
		log.Printf("failure %s", err)
		t.Fail()
	}
}

func TestInception(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	ctx := context.Background()

	innerGasLimit := 21000
	if err != nil {
		log.Fatal(err)
	}
	price, err := setup.Client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tip, err := setup.Client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	gasFeeCap, gasTipCap := env.TransactGasPriceFn(price, tip, 0, 1)
	var data []byte
	innerTx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   setup.ChainID,
			Nonce:     1,
			GasFeeCap: gasFeeCap,
			GasTipCap: gasTipCap,
			Gas:       uint64(innerGasLimit),
			To:        &setup.SubmitFromAddress,
			Value:     big.NewInt(1),
			Data:      data,
		},
	)

	// TODO: we could start a loop here
	signedInnerTx, err := setup.TransactSign(setup.TransactFromAddress, innerTx)

	encryptedInnerTx, innerIdentityPrefix, err := encrypt(ctx, *signedInnerTx, &env, setup.TransactFromAddress, 1)

	abi, err := sequencerBindings.SequencerMetaData.GetAbi()
	if err != nil {
		log.Fatal(err)
	}

	input, err := abi.Pack("submitEncryptedTransaction", env.Eon, innerIdentityPrefix, encryptedInnerTx.Marshal(), big.NewInt(int64(signedInnerTx.Gas())))
	if err != nil {
		log.Fatal(err)
	}
	price, err = setup.Client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tip, err = setup.Client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	gasFeeCap, gasTipCap = env.TransactGasPriceFn(price, tip, 0, 1)
	sequencerAddress := common.HexToAddress(SequencerContractAddress)
	middleCallMsg := ethereum.CallMsg{
		From:  setup.TransactFromAddress,
		To:    &sequencerAddress,
		Value: signedInnerTx.Cost(),
		Data:  input,
	}
	middleGasLimit, err := setup.Client.EstimateGas(ctx, middleCallMsg)
	if err != nil {
		log.Fatal(err)
	}

	middleTx := types.NewTx(
		&types.DynamicFeeTx{
			ChainID:   setup.ChainID,
			Nonce:     0,
			GasFeeCap: gasFeeCap,
			GasTipCap: gasTipCap,
			Gas:       middleGasLimit,
			To:        &sequencerAddress,
			Value:     big.NewInt(0).Sub(signedInnerTx.Cost(), signedInnerTx.Value()),
			Data:      input,
		},
	)

	signedMiddleTx, err := setup.TransactSign(setup.TransactFromAddress, middleTx)

	middleTxEncryptedMsg, middleIdentityPrefix, err := encrypt(ctx, *signedMiddleTx, &env, setup.SubmitFromAddress, 1)
	log.Println("submitting outer. Gas", signedMiddleTx.Gas())
	opts := &env.SubmitterOpts
	log.Println("submit nonce", opts.Nonce)

	// and end the loop here?
	opts.Value = big.NewInt(0).Sub(signedMiddleTx.Cost(), signedMiddleTx.Value())
	submitTx, err := setup.Sequencer.SubmitEncryptedTransaction(opts, env.Eon, middleIdentityPrefix, middleTxEncryptedMsg.Marshal(), big.NewInt(int64(signedMiddleTx.Gas())))
	if err != nil {
		log.Fatal(err)
	}
	submitReceipt, err := waitForTx(*submitTx, "outer tx", env.SubmissionWaitTimeout, setup)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(hex.EncodeToString(middleIdentityPrefix[:]))
	middleReceipt, err := waitForTx(*signedMiddleTx, "middle tx", env.InclusionWaitTimeout, setup)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(hex.EncodeToString(innerIdentityPrefix[:]))
	innerReceipt, err := waitForTx(*signedInnerTx, "inner tx", env.InclusionWaitTimeout, setup)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("inner gas", innerReceipt.GasUsed, "middle gas", middleReceipt.GasUsed, "outer gas", submitReceipt.GasUsed)
}

func TestIncorrectIdentitySuffix(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(true)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	env, err := createStressEnvironment(context.Background(), setup)
	if err != nil {
		log.Fatal("could not set up environment", err)
	}
	env.RandomIdentitySuffix = true

	err = transact(setup, &env, 1)
	if err == nil {
		log.Println("This must time out!")
		t.Fail()
	}
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

// not really a test, but useful to fix the submit account's nonce, if an earlier test failed
func TestFixNonce(t *testing.T) {
	skipCI(t)
	setup, err := createSetup(false)
	if err != nil {
		log.Fatal("could not create setup", err)
	}
	err = fixNonce(setup)
	if err != nil {
		log.Fatal(err)
	}

}

// TODO

// invalid tx

// identity suffixes
