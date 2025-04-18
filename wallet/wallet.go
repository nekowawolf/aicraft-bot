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
	CreateVoteTransaction(rpcURL, contractAddress, candidateID string, feedAmount int, chainID int64) (string, error)
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

func (w *Wallet) CreateVoteTransaction(rpcURL, contractAddress, candidateID string, feedAmount int, chainID int64) (string, error) {
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

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %v", err)
	}

	data, err := prepareVoteData(candidateID, feedAmount)
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
		gasLimit = 200000 
	}

	tx := types.NewTransaction(
		nonce,
		contractAddr,
		big.NewInt(0),    
		gasLimit,       
		gasPrice,        
		data,           
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(chainID)), w.privateKey)
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

func prepareVoteData(candidateID string, feedAmount int) ([]byte, error) {
	methodSig := crypto.Keccak256([]byte("feed(string,uint256)"))[:4]
	
	data := append(methodSig, []byte(candidateID)...)
	data = append(data, big.NewInt(int64(feedAmount)).Bytes()...)
	
	return data, nil
}