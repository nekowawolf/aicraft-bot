package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Signer interface {
	GetAddress() string
	SignMessage(message string) (string, error)
	CreateVoteTransaction(rpcURL, contractAddress, candidateID string, feedAmount int, chainID int64, requestID, requestData, userHashedMessage, integritySignature string) (string, error)
	WaitForTransactionReceipt(rpcURL, txHash string) (*types.Receipt, error)
}

type Wallet struct {
	privateKey *ecdsa.PrivateKey
}

func NewWallet(privateKeyHex string) (*Wallet, error) {
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}
	return &Wallet{privateKey: privateKey}, nil
}

func (w *Wallet) GetAddress() string {
	publicKey := w.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return ""
	}
	return crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
}

func (w *Wallet) SignMessage(message string) (string, error) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	msgHash := crypto.Keccak256Hash([]byte(msg))

	signature, err := crypto.Sign(msgHash.Bytes(), w.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %v", err)
	}

	signature[64] += 27
	return hexutil.Encode(signature), nil
}

func (w *Wallet) CreateVoteTransaction(rpcURL, contractAddress, candidateID string, feedAmount int, chainID int64, requestID, requestData, userHashedMessage, integritySignature string) (string, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return "", fmt.Errorf("failed to connect to RPC: %v", err)
	}
	defer client.Close()

	contractAddr := common.HexToAddress(contractAddress)
	fromAddress := common.HexToAddress(w.GetAddress())

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %v", err)
	}

	baseFee, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get base fee: %v", err)
	}

	priorityFee := big.NewInt(1000000000)
	maxFeePerGas := new(big.Int).Add(baseFee, priorityFee)

	data, err := prepareVoteData(candidateID, feedAmount, requestID, requestData, userHashedMessage, integritySignature)
	if err != nil {
		return "", fmt.Errorf("failed to prepare transaction data: %v", err)
	}

	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  fromAddress,
		To:    &contractAddr,
		Value: big.NewInt(0),
		Data:  data,
	})
	if err != nil {
		gasLimit = 100000 
	} else {
		gasLimit = gasLimit * 110 / 100
		if gasLimit < 100000 {
			gasLimit = 100000 
		}
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(chainID),
		Nonce:     nonce,
		GasTipCap: priorityFee,
		GasFeeCap: maxFeePerGas,
		Gas:       gasLimit,
		To:        &contractAddr,
		Value:     big.NewInt(0),
		Data:      data,
	})

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(big.NewInt(chainID)), w.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %v", err)
	}

	return signedTx.Hash().Hex(), nil
}

func (w *Wallet) WaitForTransactionReceipt(rpcURL, txHash string) (*types.Receipt, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %v", err)
	}
	defer client.Close()

	hash := common.HexToHash(txHash)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for {
		receipt, err := client.TransactionReceipt(ctx, hash)
		if err == nil {
			return receipt, nil
		}
		if err == ethereum.NotFound {
			select {
			case <-time.After(2 * time.Second):
				continue
			case <-ctx.Done():
				return nil, fmt.Errorf("timeout waiting for transaction receipt")
			}
		} else {
			return nil, fmt.Errorf("failed to get receipt: %v", err)
		}
	}
}

func prepareVoteData(candidateID string, feedAmount int, requestID, requestData, userHashedMessage, integritySignature string) ([]byte, error) {
	methodSig := crypto.Keccak256([]byte("feed(string,uint256,string,string,bytes,bytes)"))[:4]

	var data []byte
	data = append(data, methodSig...)

	baseOffset := big.NewInt(6 * 32)

	candidateIDLen := len(candidateID)
	requestIDLen := len(requestID)
	requestDataLen := len(requestData)

	var userHashedBytes []byte
	var err error
	if strings.HasPrefix(userHashedMessage, "0x") {
		userHashedBytes, err = hexutil.Decode(userHashedMessage)
	} else {
		userHashedBytes, err = hexutil.Decode("0x" + userHashedMessage)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode user hashed message: %v", err)
	}
	userHashedLen := len(userHashedBytes)

	var integrityBytes []byte
	if strings.HasPrefix(integritySignature, "0x") {
		integrityBytes, err = hexutil.Decode(integritySignature)
	} else {
		integrityBytes, err = hexutil.Decode("0x" + integritySignature)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode integrity signature: %v", err)
	}
	integrityLen := len(integrityBytes)

	offset1 := new(big.Int).Set(baseOffset)                                                 
	offset2 := new(big.Int).Add(offset1, big.NewInt(int64(32+((candidateIDLen+31)/32)*32))) 
	offset3 := new(big.Int).Add(offset2, big.NewInt(int64(32+((requestIDLen+31)/32)*32)))   
	offset4 := new(big.Int).Add(offset3, big.NewInt(int64(32+((requestDataLen+31)/32)*32))) 
	offset5 := new(big.Int).Add(offset4, big.NewInt(int64(32+((userHashedLen+31)/32)*32)))  

	data = append(data, common.LeftPadBytes(offset1.Bytes(), 32)...)

	data = append(data, common.LeftPadBytes(big.NewInt(int64(feedAmount)).Bytes(), 32)...)

	data = append(data, common.LeftPadBytes(offset2.Bytes(), 32)...)

	data = append(data, common.LeftPadBytes(offset3.Bytes(), 32)...)

	data = append(data, common.LeftPadBytes(offset4.Bytes(), 32)...)

	data = append(data, common.LeftPadBytes(offset5.Bytes(), 32)...)

	data = append(data, common.LeftPadBytes(big.NewInt(int64(candidateIDLen)).Bytes(), 32)...)
	data = append(data, common.RightPadBytes([]byte(candidateID), ((candidateIDLen+31)/32)*32)...)

	data = append(data, common.LeftPadBytes(big.NewInt(int64(requestIDLen)).Bytes(), 32)...)
	data = append(data, common.RightPadBytes([]byte(requestID), ((requestIDLen+31)/32)*32)...)

	data = append(data, common.LeftPadBytes(big.NewInt(int64(requestDataLen)).Bytes(), 32)...)
	data = append(data, common.RightPadBytes([]byte(requestData), ((requestDataLen+31)/32)*32)...)

	data = append(data, common.LeftPadBytes(big.NewInt(int64(userHashedLen)).Bytes(), 32)...)
	data = append(data, common.RightPadBytes(userHashedBytes, ((userHashedLen+31)/32)*32)...)

	data = append(data, common.LeftPadBytes(big.NewInt(int64(integrityLen)).Bytes(), 32)...)
	data = append(data, common.RightPadBytes(integrityBytes, ((integrityLen+31)/32)*32)...)

	return data, nil
}
