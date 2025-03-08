package ethereum

import (
	"context"
	"fmt"
	"strings"

	"github.com/ddomeke/rpc_proxy/internal/config"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// CheckIfAddressIsFrozen checks if an address is on the frozen accounts list
func CheckIfAddressIsFrozen(cfg *config.Config, addressToCheck string) (bool, error) {
	// Connect to the Optimism devnet
	client, err := ethclient.Dial(cfg.L1RPCURL)
	if err != nil {
		return false, fmt.Errorf("could not connect to ethereum client: %v", err)
	}
	defer client.Close()

	// Check connection
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return false, fmt.Errorf("could not get chain ID: %v", err)
	}
	fmt.Printf("Connection established, Chain ID: %s\n", chainID.String())

	frozenContractAddress := common.HexToAddress(cfg.FrozenContractAddress)
	fmt.Printf("Frozen Contract Address: %s\n", frozenContractAddress)

	checkAddress := common.HexToAddress(addressToCheck)
	fmt.Printf("Address to check: %s\n", checkAddress.Hex())

	// Check if address has contract code
	code, err := client.CodeAt(context.Background(), frozenContractAddress, nil)
	if err != nil {
		return false, fmt.Errorf("could not get contract code: %v", err)
	}
	if len(code) == 0 {
		return false, fmt.Errorf("no contract found at the specified address")
	}
	fmt.Printf("Contract code found, code length: %d bytes\n", len(code))

	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(FrozenAccountsABI))
	if err != nil {
		return false, fmt.Errorf("could not parse ABI: %v", err)
	}

	// Prepare input data for isFrozen
	input, err := parsedABI.Pack("isFrozen", checkAddress)
	if err != nil {
		return false, fmt.Errorf("could not pack input parameters: %v", err)
	}
	fmt.Printf("Request data created, length: %d bytes\n", len(input))

	// Call the contract
	ctx := context.Background()
	msg := ethereum.CallMsg{
		To:   &frozenContractAddress,
		Data: input,
	}

	fmt.Println("Calling contract...")
	output, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return false, fmt.Errorf("contract call failed: %v", err)
	}

	fmt.Printf("Call successful, output data length: %d bytes\n", len(output))
	if len(output) == 0 {
		return false, fmt.Errorf("contract returned empty output")
	}

	// Process output
	var result bool
	err = parsedABI.UnpackIntoInterface(&result, "isFrozen", output)
	if err != nil {
		return false, fmt.Errorf("could not unpack output: %v", err)
	}

	return result, nil
}
