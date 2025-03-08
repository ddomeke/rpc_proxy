package monitor

import (
	"context"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ddomeke/rpc_proxy/internal/config"
	"github.com/ddomeke/rpc_proxy/internal/eth"
	"github.com/ddomeke/rpc_proxy/internal/metrics"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const retryDelay = 5 * time.Second // Retry delay in case of errors

// Global deposit tracking
var (
	pendingDeposits      = make(map[common.Hash]*eth.DepositEvent)
	pendingDepositsMutex sync.RWMutex
)

// GetPendingDeposits returns a copy of the current pending deposits map
func GetPendingDeposits() map[common.Hash]*eth.DepositEvent {
	pendingDepositsMutex.RLock()
	defer pendingDepositsMutex.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[common.Hash]*eth.DepositEvent)
	for k, v := range pendingDeposits {
		result[k] = v
	}
	return result
}

// UpdatePendingDeposits adds a deposit to pending deposits
func UpdatePendingDeposits(deposit *eth.DepositEvent) {
	pendingDepositsMutex.Lock()
	defer pendingDepositsMutex.Unlock()
	pendingDeposits[deposit.Hash] = deposit
}

// RemovePendingDeposit removes a deposit from pending deposits
func RemovePendingDeposit(hash common.Hash) {
	pendingDepositsMutex.Lock()
	defer pendingDepositsMutex.Unlock()
	delete(pendingDeposits, hash)
}

// ListenL1DepositEvents listens for deposit events on L1
func ListenL1DepositEvents(clients *eth.Clients, cfg *config.Config, metricsCollector *metrics.Collector) {
	log.Println("[INFO] Starting L1 Deposit event listener...")

	// Listen for TransactionDeposited events at the OptimismPortal address
	portalAddress := common.HexToAddress(cfg.OptimismPortalAddress)
	depositEventSignature := common.HexToHash("0x35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f23")

	query := ethereum.FilterQuery{
		Addresses: []common.Address{portalAddress},
		Topics:    [][]common.Hash{{depositEventSignature}},
	}

	// Connect to WebSocket for event subscription
	l1Clientws, err := ethclient.Dial(cfg.L1RPCURLWs)
	if err != nil {
		log.Fatalf("[ERROR] Could not connect to L1 websocket: %v", err)
	}

	logs := make(chan types.Log)
	sub, err := l1Clientws.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("[ERROR] L1 deposit event subscription failed: %v", err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Printf("[ERROR] L1 event listening error: %v", err)
			time.Sleep(retryDelay)
			go ListenL1DepositEvents(clients, cfg, metricsCollector) // Reconnect
			return
		case logEntry := <-logs:
			// Decode TransactionDeposited event
			deposit, err := eth.DecodeDepositEvent(clients, logEntry)

			log.Printf("[DEBUG] Event topic count: %d", len(logEntry.Topics))
			for i, topic := range logEntry.Topics {
				log.Printf("[DEBUG] Topic %d: %s", i, topic.Hex())
			}
			log.Printf("[DEBUG] Data length: %d", len(logEntry.Data))

			if err != nil {
				log.Printf("[ERROR] Deposit event parsing error: %v", err)
				continue
			}

			// Check if address is frozen
			frozen, err := eth.CheckIfAddressIsFrozen(cfg, deposit.From.Hex())
			if err != nil {
				log.Printf("[ERROR] Address check error: %v", err)
			} else if frozen {
				// Block deposit from frozen account
				log.Printf("[INFO] Deposit from frozen account blocked: %s", deposit.From.Hex())
				metricsCollector.BlockedDeposits.Inc()
				continue
			}

			// Update deposit metrics
			metricsCollector.TotalDeposits.Inc()
			metricsCollector.DepositsByAccount.WithLabelValues(deposit.From.Hex()).Inc()
			metricsCollector.DepositValueTotal.Add(float64(deposit.Value.Uint64()))

			// Add ETH value to histogram
			ethValue := new(big.Float).Quo(
				new(big.Float).SetInt(deposit.Value),
				new(big.Float).SetInt64(1e18),
			)
			ethFloat, _ := ethValue.Float64()
			metricsCollector.DepositValueHistogram.Observe(ethFloat)

			// Add gas limit to histogram
			metricsCollector.DepositGasLimit.Observe(float64(deposit.GasLimit))

			// Add to pending deposits and update counter
			UpdatePendingDeposits(deposit)
			metricsCollector.DepositsPending.Set(float64(len(pendingDeposits)))

			log.Printf("[INFO] New deposit recorded: %s -> %s (%.6f ETH, gas: %d)",
				deposit.From.Hex(), deposit.To.Hex(), ethFloat, deposit.GasLimit)
		}
	}
}
