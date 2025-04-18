package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL1 = "https://api.aicraft.fun"
)

func CreateVoteOrder(token, candidateID, chainID, countryID, rpcURL, walletID string, feedAmount int) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/feeds/orders", baseURL)

	reqBody := map[string]interface{}{
		"candidateID": candidateID,
		"chainID":     chainID,
		"countryId":   countryID,
		"rpcUrl":      rpcURL,
		"walletID":    walletID,
		"feedAmount":  feedAmount,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Create order response: %s\n", string(body))

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response OrderResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Data.Order.ID == "" {
		return nil, fmt.Errorf("empty order ID received")
	}

	return &response, nil
}

func GetVoteOrder(token, orderID string) (*OrderResponse, error) {
	url := fmt.Sprintf("%s/feeds/orders/%s", baseURL, orderID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Get order response: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response OrderResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &response, nil
}

func ConfirmVoteOrder(token, orderID, txHash string) error {
	url := fmt.Sprintf("%s/feeds/orders/%s/confirm", baseURL, orderID)

	reqBody := map[string]interface{}{
		"txHash": txHash,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Confirm order response: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
