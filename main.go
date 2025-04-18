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
		log.Println("Warning: No .env file found, using system environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	printConfig(cfg)

	wallet, err := wallet.NewWallet(cfg.PrivateKey)
	if err != nil {
		log.Fatalf("❌ Failed to initialize wallet: %v", err)
	}
	fmt.Printf("🔑 Wallet address: %s\n", wallet.GetAddress())

	token, err := api.WalletSignIn(wallet)
	if err != nil {
		log.Fatalf("❌ Failed to authenticate: %v", err)
	}
	fmt.Printf("🔑 Authentication successful\n")

	fmt.Printf("🗳️ Creating vote order for candidate %s...\n", cfg.CandidateID)
	order, err := api.CreateVoteOrder(
		token,
		cfg.CandidateID,
		cfg.GetChainIDString(),
		cfg.TargetCountryID,
		cfg.RPCURL,
		cfg.WalletID,
		cfg.FeedAmount,
	)
	if err != nil {
		log.Fatalf("❌ Failed to create vote order: %v", err)
	}
	printOrderDetails(order)

	fmt.Printf("⛓ Creating blockchain transaction...\n")
	txHash, err := wallet.CreateVoteTransaction(
		cfg.RPCURL,
		order.Data.Payment.ContractAddress,
		cfg.CandidateID,
		cfg.FeedAmount,
		cfg.ChainID,
		order.Data.Order.ID,
		order.Data.Payment.Params.RequestData,
		order.Data.Payment.Params.UserHashedMessage,
		order.Data.Payment.Params.IntegritySignature,
	)
	if err != nil {
		log.Fatalf("❌ Failed to create vote transaction: %v", err)
	}
	fmt.Printf("📝 Transaction hash: %s\n", txHash)

	fmt.Printf("⏳ Waiting for transaction confirmation (timeout: 5 minutes)...\n")
	receipt, err := wallet.WaitForTransactionReceipt(cfg.RPCURL, txHash)
	if err != nil {
		log.Fatalf("❌ Failed to get transaction receipt: %v", err)
	}

	if receipt.Status != 1 {
		log.Fatalf("❌ Transaction failed: %s", txHash)
	}
	fmt.Printf("✅ Transaction confirmed in block %d\n", receipt.BlockNumber)

	fmt.Printf("✅ Confirming vote order...\n")
	if err := api.ConfirmVoteOrder(token, order.Data.Order.ID, txHash); err != nil {
		log.Fatalf("❌ Failed to confirm vote order: %v", err)
	}

	fmt.Println("\n🎉 Vote successfully submitted!")
	fmt.Printf("🔗 Transaction: %s\n", txHash)
	fmt.Printf("🗳️ Candidate: %s\n", cfg.CandidateID)
	fmt.Printf("🌎 Country: %s\n", cfg.TargetCountryID)
}

func printConfig(cfg *config.Config) {
	fmt.Println("\n⚙️ Configuration:")
	fmt.Printf("• RPC URL: %s\n", cfg.RPCURL)
	fmt.Printf("• Chain ID: %d\n", cfg.ChainID)
	fmt.Printf("• Target Country ID: %s\n", cfg.TargetCountryID)
	fmt.Printf("• Candidate ID: %s\n", cfg.CandidateID)
	fmt.Printf("• Feed Amount: %d\n", cfg.FeedAmount)
	fmt.Printf("• Delay Seconds: %d\n\n", cfg.DelaySeconds)
}

func printOrderDetails(order *api.OrderResponse) {
	fmt.Println("\n📄 Order Details:")
	fmt.Printf("• Order ID: %s\n", order.Data.Order.ID)
	fmt.Printf("• Status: %s\n", order.Data.Order.Status)
	fmt.Printf("• Contract Address: %s\n", order.Data.Payment.ContractAddress)
	fmt.Printf("• Function: %s\n", order.Data.Payment.FunctionName)
	fmt.Printf("• Feed Amount: %d\n", order.Data.Payment.Params.FeedAmount)
	fmt.Println()
}
