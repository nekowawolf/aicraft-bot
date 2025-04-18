package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/nekowawolf/aicraft-bot/api"
	"github.com/nekowawolf/aicraft-bot/config"
	"github.com/nekowawolf/aicraft-bot/wallet"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	fmt.Printf("Config Values:\n")
	fmt.Printf("Private Key: %t\n", cfg.PrivateKey != "")
	fmt.Printf("Target Country ID: %s\n", cfg.TargetCountryID)
	fmt.Printf("Candidate ID: %s\n", cfg.CandidateID)

	wallet, err := wallet.NewWallet(cfg.PrivateKey)
	if err != nil {
		log.Fatalf("❌ Failed to initialize wallet: %v", err)
	}
	fmt.Printf("🔑 Using wallet: %s\n", wallet.GetAddress())

	token, err := api.WalletSignIn(wallet)
	if err != nil {
		log.Fatalf("❌ Failed to authenticate: %v", err)
	}
	fmt.Printf("🔑 Authentication token: %s\n", token)

	fmt.Printf("🔑 Using candidate: %s\n", cfg.CandidateID)

	order, err := api.CreateVoteOrder(token, cfg.CandidateID, cfg.ChainID, cfg.TargetCountryID, cfg.RPCURL, cfg.WalletID, cfg.FeedAmount)
	if err != nil {
		log.Fatalf("❌ Failed to create vote order: %v", err)
	}
	fmt.Printf("📝 Created vote order: %s\n", order.Data.Order.ID)
	fmt.Printf("📝 Contract address: %s\n", order.Data.Payment.ContractAddress)
	fmt.Printf("📝 Feed amount: %d\n", order.Data.Payment.Params.FeedAmount)

	txHash := order.Data.Payment.Params.RequestID
	if txHash == "" {
		log.Fatalf("❌ No transaction hash found in order response")
	}
	fmt.Printf("📝 Transaction hash: %s\n", txHash)

	if err := api.ConfirmVoteOrder(token, order.Data.Order.ID, txHash); err != nil {
		log.Fatalf("❌ Failed to confirm vote order: %v", err)
	}

	fmt.Println("✅ Vote successfully submitted!")
	os.Exit(0)
}
