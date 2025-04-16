package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL = "https://api.aicraft.fun"
)

func GetSignMessage(address string) (string, error) {
	url := fmt.Sprintf("%s/auths/wallets/sign-in/message?address=%s&type=ETHEREUM_BASED", baseURL, address)
	fmt.Printf("Requesting sign message from: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get sign message: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response status: %d, body: %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if authResp.Data.Message == "" {
		return "", fmt.Errorf("empty message received from server")
	}

	return authResp.Data.Message, nil
}

func SignIn(address, signature string) (string, error) {
	url := fmt.Sprintf("%s/auths/wallets/sign-in", baseURL)

	reqBody := SignInRequest{
		Address:   address,
		Signature: signature,
		Type:      "ETHEREUM_BASED",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to sign in: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var signInResp SignInResponse
	if err := json.Unmarshal(body, &signInResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if signInResp.Data.AccessToken == "" {
		return "", fmt.Errorf("empty access token received")
	}

	return signInResp.Data.AccessToken, nil
}
