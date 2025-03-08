package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ddomeke/rpc_proxy/internal/config"
	"github.com/ddomeke/rpc_proxy/internal/eth"
	"github.com/ddomeke/rpc_proxy/internal/metrics"
	"github.com/ddomeke/rpc_proxy/pkg/utils"
	"github.com/ethereum/go-ethereum/common"
)

// MonitorL2Deposits monitors transactions on L2 and matches deposits
func MonitorL2Deposits(clients *eth.Clients, cfg *config.Config, metricsCollector *metrics.Collector) {
	log.Println("[INFO] Starting L2 deposit confirmation monitor...")

	// Last checked block
	lastCheckedBlock := uint64(0)

	for {
		// Get L2 block number
		currentBlock, err := clients.L2Client.BlockNumber(context.Background())
		if err != nil {
			log.Printf("[ERROR] Could not get L2 block number: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// If there are new blocks
		if currentBlock > lastCheckedBlock {
			// Check blocks
			for blockNum := lastCheckedBlock + 1; blockNum <= currentBlock; blockNum++ {
				checkL2BlockViaRPC(blockNum, clients, cfg, metricsCollector)
			}
			lastCheckedBlock = currentBlock
		}

		time.Sleep(2 * time.Second)
	}
}

// checkL2BlockViaRPC checks L2 block (using HTTP RPC)
func checkL2BlockViaRPC(blockNum uint64, clients *eth.Clients, cfg *config.Config, metricsCollector *metrics.Collector) {
	// Create JSON-RPC request
	blockNumHex := fmt.Sprintf("0x%x", blockNum)
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{blockNumHex, true},
	}

	requestData, _ := json.Marshal(rpcRequest)
	resp, err := http.Post(cfg.L2RPCURL, "application/json", bytes.NewBuffer(requestData))
	if err != nil {
		log.Printf("[ERROR] L2 RPC request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("[ERROR] RPC response parsing error: %v", err)
		return
	}

	// Process response
	if result, ok := response["result"].(map[string]interface{}); ok {
		// Convert block timestamp from string
		var blockTime time.Time
		if timestampHex, ok := result["timestamp"].(string); ok {
			timestamp, err := utils.HexToUint64(timestampHex)
			if err == nil {
				blockTime = time.Unix(int64(timestamp), 0)
			} else {
				log.Printf("[ERROR] Block timestamp conversion error: %v", err)
				blockTime = time.Now() // Fallback
			}
		}

		if transactions, ok := result["transactions"].([]interface{}); ok {
			for _, txData := range transactions {
				tx := txData.(map[string]interface{})
				if txHashStr, ok := tx["hash"].(string); ok {
					txHash := common.HexToHash(txHashStr)

					// Get pending deposits
					pendingMap := GetPendingDeposits()

					// Check deposit hash
					if deposit, exists := pendingMap[txHash]; exists {
						// Deposit confirmed
						confirmTime := blockTime.Sub(deposit.Timestamp)
						log.Printf("[INFO] Deposit confirmed on L2: %s (%.2f seconds)", txHash.Hex(), confirmTime.Seconds())

						// Remove from pending list
						RemovePendingDeposit(txHash)
					}
				}
			}
		}
	}
}
