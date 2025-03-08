package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/big"
	"net/http"

	"github.com/ddomeke/rpc_proxy/internal/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// proxyHandler handles JSON-RPC proxy requests
func (s *Server) proxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] JSON-RPC request received")

	// Read JSON-RPC request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request", http.StatusBadRequest)
		log.Printf("[ERROR] Could not read RPC request: %v", err)
		return
	}
	defer r.Body.Close()

	// Parse JSON-RPC request
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Printf("[ERROR] JSON parse error: %v", err)
		return
	}

	// Special handling for eth_getBlockReceipts
	if method, ok := req["method"].(string); ok && method == "eth_getBlockReceipts" {
		s.blockReceiptsHandler(w, body)
		return
	}

	// Forward all other requests directly
	resp, err := http.Post(s.config.L1RPCURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Ethereum RPC request failed", http.StatusInternalServerError)
		log.Printf("[ERROR] Ethereum RPC request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	// Forward response to client
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
	log.Println("[INFO] JSON-RPC request successfully forwarded")
}

// blockReceiptsHandler handles eth_getBlockReceipts special processing
func (s *Server) blockReceiptsHandler(w http.ResponseWriter, body []byte) {
	log.Println("[INFO] Processing eth_getBlockReceipts request...")

	// Forward request to L1 Ethereum RPC
	resp, err := http.Post(s.config.L1RPCURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Ethereum RPC request failed", http.StatusInternalServerError)
		log.Printf("[ERROR] Ethereum RPC request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Could not read response", http.StatusInternalServerError)
		log.Printf("[ERROR] Could not read response: %v", err)
		return
	}

	// Parse response as JSON
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(respBody, &jsonResponse); err != nil {
		http.Error(w, "Could not parse response as JSON", http.StatusInternalServerError)
		log.Printf("[ERROR] JSON parse error: %v", err)
		return
	}

	// Check for TransactionDeposited events in logs
	if result, ok := jsonResponse["result"].([]interface{}); ok {
		for _, tx := range result {
			txMap := tx.(map[string]interface{})
			if logs, ok := txMap["logs"].([]interface{}); ok {
				filteredLogs := []interface{}{}
				for _, logEntry := range logs {
					logMap := logEntry.(map[string]interface{})

					// If event is TransactionDeposited, check from address
					if logMap["topics"] != nil {
						topics := logMap["topics"].([]interface{})
						if len(topics) > 0 && topics[0] == "0x35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f23" {
							// From address
							if len(topics) > 1 {
								fromAddrHex := topics[1].(string)
								fromAddress := common.HexToAddress(fromAddrHex)

								frozen, err := eth.CheckIfAddressIsFrozen(s.config, fromAddress.Hex())
								if err != nil {
									log.Printf("[ERROR] Frozen address check error: %v", err)
									continue
								}
								if frozen {
									log.Printf("[INFO] Frozen account found: %s", fromAddress.Hex())

									s.metricsCollector.BlockedDeposits.WithLabelValues(fromAddress.Hex()).Inc()
									continue // Filter out this log
								}

								// Decode TransactionDeposited event data
								if blockNumberHex, ok := logMap["blockNumber"].(string); ok {
									_, _ = hexutil.DecodeUint64(blockNumberHex)

									// Decode log data
									data, _ := hexutil.Decode(logMap["data"].(string))

									var value *big.Int
									var gasLimit uint64

									// Simplified data decoding (full decode possible)
									if len(data) >= 96 { // At least 3 32-byte parameters
										value = new(big.Int).SetBytes(data[0:32])
										gasLimitBytes := data[32:64]
										gasLimit = new(big.Int).SetBytes(gasLimitBytes).Uint64()

										// Update metrics
										ethValue := new(big.Float).Quo(
											new(big.Float).SetInt(value),
											new(big.Float).SetInt64(1e18),
										)
										ethFloat, _ := ethValue.Float64()

										log.Printf("[INFO] Deposit monitored: %s -> Value: %.6f ETH, Gas: %d",
											fromAddress.Hex(), ethFloat, gasLimit)

										s.metricsCollector.DepositValueHistogram.Observe(ethFloat)
									}
								}
							}
						}
					}
					filteredLogs = append(filteredLogs, logMap)
				}
				txMap["logs"] = filteredLogs
			}
		}
	}

	// Forward updated JSON to client
	filteredResponse, _ := json.Marshal(jsonResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(filteredResponse)
	log.Println("[INFO] Frozen accounts filtered and response forwarded.")
}
