package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ABI Definitions
const (
	// FrozenAccountsABI - Complete ABI definition of the contract
	FrozenAccountsABI = `[
		{
			"inputs": [],
			"stateMutability": "nonpayable",
			"type": "constructor"
		},
		{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"internalType": "address",
					"name": "account",
					"type": "address"
				}
			],
			"name": "AccountFrozen",
			"type": "event"
		},
		{
			"inputs": [
				{
					"internalType": "address",
					"name": "_account",
					"type": "address"
				}
			],
			"name": "freezeAccount",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		},
		{
			"inputs": [
				{
					"internalType": "address",
					"name": "",
					"type": "address"
				}
			],
			"name": "frozen",
			"outputs": [
				{
					"internalType": "bool",
					"name": "",
					"type": "bool"
				}
			],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [
				{
					"internalType": "address",
					"name": "_account",
					"type": "address"
				}
			],
			"name": "isFrozen",
			"outputs": [
				{
					"internalType": "bool",
					"name": "",
					"type": "bool"
				}
			],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [],
			"name": "owner",
			"outputs": [
				{
					"internalType": "address",
					"name": "",
					"type": "address"
				}
			],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [
				{
					"internalType": "address",
					"name": "_account",
					"type": "address"
				}
			],
			"name": "unfreezeAccount",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`

	// OptimismPortalABI - Only the TransactionDeposited event definition
	OptimismPortalABI = `[
		{
			"anonymous": false,
			"inputs": [
				{
					"indexed": true,
					"internalType": "address",
					"name": "from",
					"type": "address"
				},
				{
					"indexed": true,
					"internalType": "address",
					"name": "to",
					"type": "address"
				},
				{
					"indexed": false,
					"internalType": "uint256",
					"name": "value",
					"type": "uint256"
				},
				{
					"indexed": false,
					"internalType": "uint64",
					"name": "gasLimit",
					"type": "uint64"
				},
				{
					"indexed": false,
					"internalType": "bool",
					"name": "isCreation",
					"type": "bool"
				},
				{
					"indexed": false,
					"internalType": "bytes",
					"name": "data",
					"type": "bytes"
				},
				{
					"indexed": true,
					"internalType": "bytes32",
					"name": "depositHash",
					"type": "bytes32"
				}
			],
			"name": "TransactionDeposited",
			"type": "event"
		}
	]`
)

// DepositEvent - Data structure for the deposit event
type DepositEvent struct {
	From       common.Address
	To         common.Address
	Value      *big.Int
	GasLimit   uint64
	IsCreation bool
	Data       []byte
	Hash       common.Hash
	BlockNum   uint64
	TxIndex    uint
	LogIndex   uint
	Timestamp  time.Time
}

// DecodeDepositEvent decodes the TransactionDeposited event
func DecodeDepositEvent(clients *Clients, log types.Log) (*DepositEvent, error) {
	// Expect at least 3 topics (event signature, from, to)
	if len(log.Topics) < 3 {
		return nil, fmt.Errorf("insufficient number of topics")
	}

	var event DepositEvent

	// Get indexed fields from topics
	event.From = common.HexToAddress(log.Topics[1].Hex())
	event.To = common.HexToAddress(log.Topics[2].Hex())

	// Set hash to default value or calculate (if available)
	if len(log.Topics) > 3 {
		event.Hash = common.HexToHash(log.Topics[3].Hex())
	} else {
		// If no hash, use tx hash
		event.Hash = log.TxHash
	}

	// Try to decode other fields from data
	if len(log.Data) > 0 {
		// Data format may vary, use simplified approach
		if len(log.Data) >= 96 { // At least 3 32-byte parameters
			// First 32 bytes: value
			event.Value = new(big.Int).SetBytes(log.Data[0:32])
			// Second 32 bytes: gasLimit
			gasLimitBytes := log.Data[32:64]
			event.GasLimit = new(big.Int).SetBytes(gasLimitBytes).Uint64()
		}
	}

	// Add block information
	event.BlockNum = log.BlockNumber
	event.TxIndex = log.TxIndex
	event.LogIndex = log.Index

	// Get block timestamp
	header, err := clients.L1Client.HeaderByNumber(context.Background(), big.NewInt(int64(log.BlockNumber)))
	if err == nil { // If no error, add timestamp
		event.Timestamp = time.Unix(int64(header.Time), 0)
	} else {
		event.Timestamp = time.Now() // Use current time as fallback
	}

	return &event, nil
}
