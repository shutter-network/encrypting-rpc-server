package server_test

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/core/types"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/shutter-network/encrypting-rpc-server/contracts"
	rpc "github.com/shutter-network/encrypting-rpc-server/rpc"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

func backendTest(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(ServerURL)
	if err != nil {
		t.Fail()
	}
	err, log := captureOutput(func() error {
		_, err := client.ChainID(ctx)
		return err
	})
	if err != nil {
		t.Fail()
	}
	if !strings.Contains(log, "dispatching to backend") {
		t.Fail()
	}
}

func processorTest(t *testing.T) {
	slot := uint64(0)
	rpc.ComputeSlot = func(blockTimestamp uint64) (*uint64, error) { return &slot, nil }
	rpc.EpochComputer = func(epochIDBytes []byte) *shcrypto.EpochID { return TestEpochID }
	ctx := context.Background()
	fromAddress := common.HexToAddress(TxFromAddress)
	toAddress := common.HexToAddress(TxToAddress)
	client, err := ethclient.Dial(ServerURL)
	if err != nil {
		t.Fail()
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get chain id")
	}

	privateKey, err := crypto.HexToECDSA(TxPrivKey)
	if err != nil {
		t.Fail()
	}

	contractInfo, err := getContractData()
	if err != nil {
		log.Fatal().Err(err).Msg("can not get contract info")
	}

	sequencerContract, err := contracts.NewSequencerContract(contractInfo["Sequencer"], client)
	if err != nil {
		log.Fatal().Err(err).Msg("can not get sequencer contract")
	}

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		t.Fail()
	}
	amount := big.NewInt(int64(math.Pow(10, 18)))
	gasLimit := uint64(21000)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fail()
	}
	err, log := captureOutput(func() error {
		tx := txtypes.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)
		signer := types.NewEIP155Signer(chainId)
		signedTx, err := types.SignTx(tx, signer, privateKey)
		if err != nil {
			return err
		}
		log.Info().Interface("tx", signedTx).Msg("test")
		buf := new(bytes.Buffer)
		ts := txtypes.Transactions{tx}
		ts.EncodeIndex(0, buf)
		rawTx := hexutil.Encode(buf.Bytes())
		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			log.Fatal().Err(err).Msg("can not send")
		}
		opts := bind.FilterOpts{Start: 0}
		it, err := sequencerContract.FilterTransactionSubmitted(&opts)
		if err != nil {
			log.Fatal().Err(err).Msg("can not get events")
		}
		for {
			if it.Event != nil {
				encryptedTx := it.Event.EncryptedTransaction
				identity := rpc.ComputeIdentity(it.Event.IdentityPrefix[:], it.Event.Sender)
				identityPreimage := identitypreimage.IdentityPreimage(identity.Marshal())
				message := shcrypto.EncryptedMessage{}
				err := message.Unmarshal(encryptedTx)
				if err != nil {
					log.Fatal().Err(err).Msg("can not unmarshall encrypted tx")
				}
				decryptKey := TestKeygen.EpochSecretKey(identityPreimage)
				fmt.Println("decrypt", decryptKey.Marshal())
				decrypted_tx, err := message.Decrypt(decryptKey)
				if err != nil {
					log.Fatal().Err(err).Msg("can not decrypt encrypted tx")
				}
				if hexutil.Encode(decrypted_tx) != rawTx {
					t.Fail()
				}
				break
			}
			ok := it.Next()
			if !ok {
				break
			}
		}
		return err
	})
	fmt.Println(log)

	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	if !strings.Contains(log, "dispatching to processor") {
		t.Fail()
	}
}

func TestServer(t *testing.T) {
	proc := setupServer()

	// t.Run("backend test", backendTest)
	t.Run("processor test", processorTest)
	t.Cleanup(func() {
		proc.Kill()
	})
}
