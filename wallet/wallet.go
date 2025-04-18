package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Signer interface {
	GetAddress() string
	SignMessage(message string) (string, error)
	GetTransactionHash(rpcURL, contractAddress string, amount int) (string, error)
}

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	client     *ethclient.Client
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

func (w *Wallet) GetTransactionHash(rpcURL, contractAddress string, amount int) (string, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ethereum client: %v", err)
	}
	defer client.Close()

	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get latest block: %v", err)
	}

	transactions := block.Transactions()
	if len(transactions) == 0 {
		return "", fmt.Errorf("no transactions found in the latest block")
	}

	for _, tx := range transactions {
		if tx.To() != nil && tx.To().Hex() == contractAddress {
			from, err := client.TransactionSender(context.Background(), tx, block.Hash(), 0)
			if err != nil {
				continue
			}
			if from.Hex() == w.GetAddress() {
				return tx.Hash().Hex(), nil
			}
		}
	}

	return "", fmt.Errorf("no matching transaction found")
}
