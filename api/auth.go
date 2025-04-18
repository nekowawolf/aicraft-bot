package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nekowawolf/aicraft-bot/wallet"
)

const (
	baseURL = "https://api.aicraft.fun"
)

func WalletSignIn(signer wallet.Signer) (string, error) {
	address := signer.GetAddress()
	
	message, err := getSignMessage(address)
	if err != nil {
		return "", fmt.Errorf("failed to get sign message: %v", err)
	}

	signature, err := signer.SignMessage(message)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %v", err)
	}

	token, err := authenticate(address, message, signature)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate: %v", err)
	}

	return token, nil
}

func getSignMessage(walletAddress string) (string, error) {
	url := fmt.Sprintf("%s/auths/wallets/sign-in/message?address=%s&type=ETHEREUM_BASED", baseURL, walletAddress)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Data.Message == "" {
		return "", fmt.Errorf("empty message received")
	}

	return response.Data.Message, nil
}

func authenticate(walletAddress, message, signature string) (string, error) {
    authReq := map[string]string{
        "address":   walletAddress,
        "message":   message,
        "signature": signature,
        "type":      "ETHEREUM_BASED",
    }

    jsonBody, err := json.Marshal(authReq)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %v", err)
    }

    resp, err := http.Post(
        fmt.Sprintf("%s/auths/wallets/sign-in", baseURL),
        "application/json",
        bytes.NewBuffer(jsonBody),
    )
    if err != nil {
        return "", fmt.Errorf("HTTP request failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
    }

    var authResponse struct {
        Data struct {
            Token string `json:"token"`
        } `json:"data"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
        return "", fmt.Errorf("failed to decode response: %v", err)
    }

    if authResponse.Data.Token == "" {
        return "", fmt.Errorf("empty token received")
    }

    return authResponse.Data.Token, nil
}
