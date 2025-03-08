package config

import (
	"fmt"
	"os"
)

// Config holds all the configuration settings for the application
type Config struct {
	// Ethereum RPC URLs
	L1RPCURL    string
	L1RPCURLWs  string
	L2RPCURL    string
	ProxyPort   string
	MetricsPort string

	// Contract addresses
	FrozenContractAddress string
	OptimismPortalAddress string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	l1RPC := os.Getenv("L1_RPC_URL")
	if l1RPC == "" {
		return nil, fmt.Errorf("L1_RPC_URL is not set")
	}

	l1RPCWs := os.Getenv("L1_RPC_URL_WS")
	if l1RPCWs == "" {
		return nil, fmt.Errorf("L1_RPC_URL_WS is not set")
	}

	l2RPC := os.Getenv("L2_RPC_URL")
	if l2RPC == "" {
		return nil, fmt.Errorf("L2_RPC_URL is not set")
	}

	frozenContract := os.Getenv("FROZEN_CONTRACT_ADDRESS")
	if frozenContract == "" {
		return nil, fmt.Errorf("FROZEN_CONTRACT_ADDRESS is not set")
	}

	portalAddress := os.Getenv("OPTIMISM_PORTAL_ADDRESS")
	if portalAddress == "" {
		return nil, fmt.Errorf("OPTIMISM_PORTAL_ADDRESS is not set")
	}

	proxyPort := os.Getenv("PROXY_PORT")
	if proxyPort == "" {
		proxyPort = "8545" // Default port
		fmt.Println("[WARN] PROXY_PORT not set, using default: 8545")
	}

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9100" // Default port
		fmt.Println("[WARN] METRICS_PORT not set, using default: 9100")
	}

	return &Config{
		L1RPCURL:             l1RPC,
		L1RPCURLWs:           l1RPCWs,
		L2RPCURL:             l2RPC,
		ProxyPort:            proxyPort,
		MetricsPort:          metricsPort,
		FrozenContractAddress: frozenContract,
		OptimismPortalAddress: portalAddress,
	}, nil
}
