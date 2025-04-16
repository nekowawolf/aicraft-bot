package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	PrivateKey      string
	RPCURL          string
	ChainID         int64
	TargetCountry   string
	TargetCountryID string
	MaxAttempts     int
	DelaySeconds    int
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		RPCURL:          "https://testnet-rpc.monad.xyz",
		ChainID:         10143,
		PrivateKey:      getEnv("PRIVATE_KEY", ""),
		TargetCountry:   getEnv("TARGET_COUNTRY", "ID"), 
		TargetCountryID: getEnv("TARGET_COUNTRY_ID", ""),
		MaxAttempts:     getEnvAsInt("MAX_ATTEMPTS", 3),
		DelaySeconds:    getEnvAsInt("DELAY_SECONDS", 5),
	}

	if cfg.PrivateKey == "" {
		return nil, fmt.Errorf("PRIVATE_KEY is required")
	}

	if cfg.TargetCountryID == "" {
		return nil, fmt.Errorf("TARGET_COUNTRY_ID is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}