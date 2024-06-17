package stress

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"log"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	shopContractBindings "github.com/shutter-network/shop-contracts/bindings"
	"github.com/shutter-network/shutter/shlib/shcrypto"
	"golang.org/x/exp/maps"
)

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

// contains the context for the current stress test to create transactions
type StressEnvironment struct {
	TransacterOpts        bind.TransactOpts
	TransactStartingNonce *big.Int
	TransactGasPriceFn    GasPriceFn
	TransactGasLimitFn    GasLimitFn
	InclusionWaitTimeout  time.Duration
	InclusionConstraints  ConstraintFn
	SubmitterOpts         bind.TransactOpts
	SubmitStartingNonce   *big.Int
	SubmissionWaitTimeout time.Duration
	Eon                   uint64
	EonPublicKey          *shcrypto.EonPublicKey
	WaitOnEverySubmit     bool
	IdentityPrefixes      []shcrypto.Block
}

type GasFeeCap *big.Int
type GasTipCap *big.Int

type GasLimitFn func(value *big.Int, data []byte, toAddress *common.Address, i int, count int) uint64

type GasPriceFn func(suggestedGasTipCap *big.Int, suggestedGasPrice *big.Int, i int, count int) (GasFeeCap, GasTipCap)

type ConstraintFn func(inclusions []*types.Receipt) error

func waitForTx(tx types.Transaction, description string, timeout time.Duration, setup StressSetup) (*types.Receipt, error) {
	log.Println("waiting for "+description+" ", tx.Hash().Hex())
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	receipt, err := bind.WaitMined(ctx, setup.Client, &tx)
	if err != nil {
		return nil, fmt.Errorf("error on WaitMined %s", err)
	}
	log.Println("status", receipt.Status, "block", receipt.BlockNumber)
	if receipt.Status != 1 {
		return nil, fmt.Errorf("included tx failed")
	}
	return receipt, nil
}

func countAndLog(receipts []*types.Receipt) error {
	c := map[string]uint16{}
	g := map[string]uint64{}
	for _, receipt := range receipts {
		n := receipt.BlockNumber.Text(10)
		c[n]++
		g[n] += receipt.GasUsed
	}
	log.Println("block\ttxs\tgas used")
	keys := maps.Keys(c)
	sort.Strings(keys)
	for _, n := range keys {
		log.Println(n, "\t", c[n], "\t", g[n])
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

func fixNonce(setup StressSetup) error {
	value := big.NewInt(1) // 1 wei
	gasLimit := uint64(21000)

	var data []byte
	headNonce, err := setup.Client.NonceAt(context.Background(), setup.SubmitFromAddress, nil)
	if err != nil {
		return err
	}
	log.Println("HeadNonce", headNonce)

	pendingNonce, err := setup.Client.PendingNonceAt(context.Background(), setup.SubmitFromAddress)
	if err != nil {
		return err
	}
	log.Println("PendingNonce", pendingNonce)
	var txs []types.Transaction
	for i := uint64(0); i < pendingNonce-headNonce; i++ {
		headNonce, err := setup.Client.NonceAt(context.Background(), setup.SubmitFromAddress, nil)
		if err != nil {
			return err
		}
		log.Println("HeadNonce", headNonce, "Pending", pendingNonce, "current", headNonce+i, "i", i)

		gasPrice, err := setup.Client.SuggestGasPrice(context.Background())
		if err != nil {
			return err
		}
		gasPrice = gasPrice.Add(gasPrice, gasPrice)
		tx := types.NewTransaction(headNonce+i, setup.SubmitFromAddress, value, gasLimit, gasPrice, data)
		signedTx, err := setup.SubmitSign(setup.SubmitFromAddress, tx)
		if err != nil {
			return err
		}
		err = setup.Client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Println("error on send", err)
		}
		log.Println("sent nonce fix tx", signedTx.Hash().Hex(), "to", setup.SubmitFromAddress)
		txs = append(txs, *signedTx)
	}

	log.Println("waiting for tx")
	for _, signedTx := range txs {
		_, err = bind.WaitMined(context.Background(), setup.Client, &signedTx)
		if err != nil {
			log.Println("error on wait", err)
		}
		headNonce, err := setup.Client.NonceAt(context.Background(), setup.SubmitFromAddress, nil)
		if err != nil {
			return err
		}
		log.Println("HeadNonce", headNonce, "Pending", pendingNonce)
	}
	return err
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
