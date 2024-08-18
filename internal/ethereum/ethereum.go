package ethereum

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumClient wraps the functionality to interact with an Ethereum contract
type EthereumClient struct {
	client          *ethclient.Client
	contractABI     *abi.ABI
	contractAddress common.Address
	topicToFilter   string
}

// NewEthereumClient creates a new EthereumClient instance
func NewEthereumClient(
	providerURL,
	contractAddress,
	contractABIPath string,
	topicToFilter string,
) (*EthereumClient, error) {
	// Connect to the Ethereum client
	client, err := ethclient.Dial(providerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %w", err)
	}

	// Parse the ABI
	abiFileContent, err := os.ReadFile(contractABIPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract ABI file: %w", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiFileContent)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	// Define the contract address
	contractAddr := common.HexToAddress(contractAddress)

	return &EthereumClient{
		client:          client,
		contractABI:     &parsedABI,
		contractAddress: contractAddr,
		topicToFilter:   topicToFilter,
	}, nil
}

// FetchLogs retrieves and processes logs from the contract
func (ec *EthereumClient) FetchLogs(from, to int64) error {
	log.Println("Fetching logs...")
	// Create a filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{ec.contractAddress},
		Topics:    [][]common.Hash{},
		FromBlock: big.NewInt(from),
		ToBlock:   big.NewInt(to),
	}

	// Fetch logs
	logs, err := ec.client.FilterLogs(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to retrieve logs: %w", err)
	}

	log.Printf("Fetched %d logs\n", len(logs))

	// Process and print logs
	for _, vLog := range logs {

		fmt.Printf("Log Block Number: %d\n", vLog.BlockNumber)
		fmt.Printf("Log Index: %d\n", vLog.Index)
		fmt.Printf("Log Address: %s\n", vLog.Address.Hex())
		fmt.Printf("Log Data: %s\n", common.Bytes2Hex(vLog.Data)) // Use Bytes2Hex here
		fmt.Printf("Log Topics: %v\n", vLog.Topics)

		for _, topic := range vLog.Topics {
			if topic.Hex() == ec.topicToFilter {
				fmt.Println("Matched topic!")
			}
		}

		// Decode log data if needed
		event := struct {
			Value *big.Int
		}{}
		err := ec.contractABI.UnpackIntoInterface(&event, "ValueChanged", vLog.Data)
		if err != nil {
			log.Printf("Failed to unpack log data: %v", err)
		} else {
			fmt.Printf("Event Value: %s\n", event.Value.String())
		}
	}

	log.Println("End of fetching logs...")

	return nil
}
