package ethereum

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Replace with your contract address and ABI
const contractAddress = "0xYourContractAddressHere"
const contractABI = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"value","type":"uint256"}],"name":"ValueChanged","type":"event"}]`

func main() {
	// Connect to the Ethereum client
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR_INFURA_PROJECT_ID")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	// Define the contract address
	contractAddr := common.HexToAddress(contractAddress)

	// Create a filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr},
		Topics:    [][]common.Hash{},
	}

	// Fetch logs
	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatalf("Failed to retrieve logs: %v", err)
	}

	// Process and print logs
	for _, vLog := range logs {
		fmt.Printf("Log Block Number: %d\n", vLog.BlockNumber)
		fmt.Printf("Log Index: %d\n", vLog.Index)
		fmt.Printf("Log Address: %s\n", vLog.Address.Hex())
		fmt.Printf("Log Data: %s\n", vLog.Data.Hex())
		fmt.Printf("Log Topics: %v\n", vLog.Topics)

		// Decode log data if needed
		event := struct {
			Value *big.Int
		}{}
		err := parsedABI.UnpackIntoInterface(&event, "ValueChanged", vLog.Data)
		if err != nil {
			log.Printf("Failed to unpack log data: %v", err)
		} else {
			fmt.Printf("Event Value: %s\n", event.Value.String())
		}
	}
}
