package main

import (
	"log"
	"path/filepath"

	"github.com/ddomeke/rpc_proxy/internal/config"
	"github.com/ddomeke/rpc_proxy/internal/eth"
	"github.com/ddomeke/rpc_proxy/internal/metrics"
	"github.com/ddomeke/rpc_proxy/internal/monitor"
	"github.com/ddomeke/rpc_proxy/internal/proxy"
	"github.com/ddomeke/rpc_proxy/pkg/utils"
)

func main() {
	// Initialize logging system
	logFile := utils.InitLogger()
	defer logFile.Close()

	// Load environment variables
	envPath := filepath.Join("..", "..", ".env")
	err := utils.LoadEnvFile(envPath)
	if err != nil {
		log.Fatalf("Could not load .env file: %v", err)
	}

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Ethereum clients
	ethClients, err := eth.InitClients(cfg)
	if err != nil {
		log.Fatalf("[ERROR] Could not initialize clients: %v", err)
	}

	// Initialize metrics
	metricsCollector := metrics.NewCollector()

	// Start Prometheus metrics server
	go metrics.StartServer(cfg.MetricsPort)

	// Start listening for L1 deposit events
	go monitor.ListenL1DepositEvents(ethClients, cfg, metricsCollector)

	// Monitor L2 deposit confirmations
	go monitor.MonitorL2Deposits(ethClients, cfg, metricsCollector)

	// Start JSON-RPC Proxy
	proxyServer := proxy.NewServer(cfg, ethClients, metricsCollector)
	if err := proxyServer.Start(); err != nil {
		log.Fatalf("[ERROR] Failed to start proxy server: %v", err)
	}
}
