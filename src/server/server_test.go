package server_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/shutter-network/encrypting-rpc-server/test"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ethereum/go-ethereum/core/types"
	txtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	sequencerBindings "github.com/shutter-network/gnosh-contracts/gnoshcontracts/sequencer"
	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/identitypreimage"
	"github.com/shutter-network/shutter/shlib/shcrypto"
)

func backendTest(t *testing.T) {
	ctx := context.Background()
	client, err := ethclient.Dial(test.ServerURL)
	if err != nil {
		log.Info().Err(err).Msg("can not dial to server url")
		t.FailNow()
	}
	err, logs := test.CaptureOutput(func() error {
		_, err := client.ChainID(ctx)
		return err
	})
	if err != nil {
		log.Info().Err(err).Msg("can not get chain id")
		t.FailNow()
	}
	if !strings.Contains(logs, "dispatching to backend") {
		log.Info().Msg("dispatch message not found for backend")
		t.FailNow()
	}
}

func processorTest(t *testing.T) {
	ctx := context.Background()
	fromAddress := common.HexToAddress(test.TxFromAddress)
	toAddress := common.HexToAddress(test.TxToAddress)
	client, err := ethclient.Dial(test.ServerURL)
	if err != nil {
		log.Info().Err(err).Msg("can not dial to server url")
		t.FailNow()
	}
	chainId, err := client.ChainID(ctx)
	if err != nil {
		log.Info().Err(err).Msg("can not get chain id")
		t.FailNow()
	}

	privateKey, err := crypto.HexToECDSA(test.TxPrivKey)
	if err != nil {
		log.Info().Err(err).Msg("can not cast to private key")
		t.FailNow()
	}

	contractInfo, err := test.GetContractData()
	if err != nil {
		log.Info().Err(err).Msg("can not get contract data")
		t.FailNow()
	}

	sequencerContract, err := sequencerBindings.NewSequencer(contractInfo["Sequencer"], client)
	if err != nil {
		log.Info().Err(err).Msg("can not get sequencer contract")
		t.FailNow()
	}

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Info().Err(err).Msg("can not get pending nonce")
		t.FailNow()
	}
	amount := big.NewInt(int64(math.Pow(10, 18)))
	gasLimit := uint64(21000)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Info().Err(err).Msg("can not get gas price")
		t.FailNow()
	}
	err, logs := test.CaptureOutput(func() error {
		tx := txtypes.NewTransaction(nonce, toAddress, amount, gasLimit, gasPrice, nil)
		signer := types.NewEIP155Signer(chainId)
		signedTx, err := types.SignTx(tx, signer, privateKey)
		if err != nil {
			return err
		}
		buf := new(bytes.Buffer)
		err = signedTx.EncodeRLP(buf)
		if err != nil {
			return err
		}
		rawTx := hexutil.Encode(buf.Bytes())
		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			return err
		}
		opts := bind.FilterOpts{Start: 0}
		it, err := sequencerContract.FilterTransactionSubmitted(&opts, []uint64{}, []uint64{})
		if err != nil {
			return err
		}
		for {
			if it.Event != nil {
				encryptedTx := it.Event.EncryptedTransaction
				preim := it.Event.IdentityPrefix[:]
				preim = append(preim, it.Event.Sender.Bytes()...)
				identityPreimage := identitypreimage.IdentityPreimage(preim)
				message := &shcrypto.EncryptedMessage{}
				err := message.Unmarshal(encryptedTx)
				if err != nil {
					return err
				}
				decryptKey := test.TestKeygen.EpochSecretKey(identityPreimage)
				decryptedTx, err := message.Decrypt(decryptKey)
				if err != nil {
					return err
				}
				decryptHex := hexutil.Encode(decryptedTx)
				if decryptHex != rawTx {
					return errors.New("decrypted tx is different")
				}
				break
			}
			ok := it.Next()
			if !ok {
				break
			}
		}
		return nil
	})
	fmt.Println(logs)

	if err != nil {
		log.Info().Err(err).Msg("process failed")
		t.FailNow()
	}
	if !strings.Contains(logs, "dispatching to processor") {
		log.Info().Msg("dispatch message not found in logs for processor")
		t.FailNow()
	}
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	err := test.SetupServer(ctx, t)
	if err != nil {
		log.Info().Err(err).Msg("setupServer failed")
		t.FailNow()
	}
	log.Info().Msg("setup successful")
	backendSuccess := t.Run("backend test", backendTest)
	log.Info().Bool("backend", backendSuccess).Msg("return")

	processorSuccess := t.Run("processor test", processorTest)
	log.Info().Bool("process", processorSuccess).Msg("return")
}
