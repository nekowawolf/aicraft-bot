package wallet

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewWallet(privateKeyHex string) (*Wallet, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Wallet{
		privateKey: privateKey,
		address:    address,
	}, nil
}

func (w *Wallet) GetAddress() string {
	return w.address.Hex()
}

func (w *Wallet) SignMessage(message string) (string, error) {
	messageBytes := []byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message))

	hash := crypto.Keccak256(messageBytes)

	signatureBytes, err := crypto.Sign(hash, w.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %v", err)
	}

	signatureBytes[64] += 27

	return hexutil.Encode(signatureBytes), nil
}
