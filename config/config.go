package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	PrivateKey      string `envconfig:"PRIVATE_KEY" required:"true"`
	RPCURL          string `envconfig:"RPC_URL" default:"https://testnet-rpc.monad.xyz"`
	WalletID        string `envconfig:"WALLET_ID" required:"true"`
	ChainID         int64  `envconfig:"CHAIN_ID" default:"10143"` // Diubah menjadi int64
	TargetCountry   string `envconfig:"TARGET_COUNTRY"`
	TargetCountryID string `envconfig:"TARGET_COUNTRY_ID" required:"true"`
	CandidateID     string `envconfig:"CANDIDATE_ID" required:"true"`
	FeedAmount      int    `envconfig:"FEED_AMOUNT" default:"1"`
	DelaySeconds    int    `envconfig:"DELAY_SECONDS" default:"5"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: No .env file found, using environment variables")
	}

	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process env vars: %v", err)
	}

	cfg.PrivateKey = strings.TrimSpace(cfg.PrivateKey)
	cfg.WalletID = strings.TrimSpace(cfg.WalletID)
	cfg.TargetCountryID = strings.TrimSpace(cfg.TargetCountryID)
	cfg.CandidateID = strings.TrimSpace(cfg.CandidateID)

	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("PRIVATE_KEY is required")
	}
	if cfg.WalletID == "" {
		return nil, fmt.Errorf("WALLET_ID is required")
	}
	if cfg.TargetCountryID == "" {
		return nil, fmt.Errorf("TARGET_COUNTRY_ID is required")
	}
	if cfg.CandidateID == "" {
		return nil, fmt.Errorf("CANDIDATE_ID is required")
	}

	if cfg.RPCURL == "" {
		cfg.RPCURL = "https://testnet-rpc.monad.xyz"
	}

	if cfg.ChainID == 0 {
		cfg.ChainID = 10143 
	}

	return &cfg, nil
}

func (c *Config) GetChainIDString() string {
	return strconv.FormatInt(c.ChainID, 10)
}