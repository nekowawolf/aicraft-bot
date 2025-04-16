package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/nekowawolf/aicraft-bot/api"
	"github.com/nekowawolf/aicraft-bot/config"
	"github.com/nekowawolf/aicraft-bot/wallet"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	wallet, err := wallet.NewWallet(cfg.PrivateKey)
	if err != nil {
		log.Fatalf("Failed to initialize wallet: %v", err)
	}
	address := wallet.GetAddress()
	fmt.Printf("Using wallet: %s\n", address)

	message, err := api.GetSignMessage(address)
	if err != nil {
		log.Fatalf("Failed to get sign message: %v", err)
	}
	fmt.Printf("Sign message: %s\n", message)

	signature, err := wallet.SignMessage(message)
	if err != nil {
		log.Fatalf("Failed to sign message: %v", err)
	}

	token, err := api.SignIn(address, signature)
	if err != nil {
		log.Fatalf("Failed to sign in: %v", err)
	}
	fmt.Printf("Successfully authenticated\n")

	orderID, err := api.CreateVoteOrder(token, cfg.TargetCountryID)
	if err != nil {
		log.Fatalf("Failed to create vote order: %v", err)
	}
	fmt.Printf("Created vote order: %s\n", orderID)

	if err := api.ConfirmVoteOrder(token, orderID); err != nil {
		log.Fatalf("Failed to confirm vote order: %v", err)
	}

	fmt.Println("Vote successfully submitted!")
}
