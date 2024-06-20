package test_data

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

func GenerateKeyPair() (*ecdsa.PrivateKey, common.Address, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, common.Address{}, err
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, common.Address{}, err
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return privateKey, fromAddress, nil
}

func Tx(privateKey *ecdsa.PrivateKey, nonce uint64, chainID *big.Int) (string, *types.Transaction, error) {
	toAddress := common.HexToAddress("0xC0058BdcC93EaA1afd468f06A26394E2d80c8f01")
	value := big.NewInt(120000000000) // in wei (0.12 eth)
	gasLimit := uint64(21000)         // in units
	maxPriorityFeePerGas := big.NewInt(2000000000)
	maxFeePerGas := big.NewInt(20000000000)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		To:        &toAddress,
		Value:     value,
		Gas:       gasLimit,
		GasFeeCap: maxFeePerGas,
		GasTipCap: maxPriorityFeePerGas,
	})

	signer := types.NewLondonSigner(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", nil, err
	}

	rawTxBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return "", nil, err
	}
	rawTx := "0x" + common.Bytes2Hex(rawTxBytes)

	return rawTx, signedTx, nil
}
