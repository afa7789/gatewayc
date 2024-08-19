// client/client.go
package client

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/afa7789/gatewayc/internal/boltdb"
	"github.com/afa7789/gatewayc/internal/domain"
	"github.com/afa7789/gatewayc/internal/ethereum"
)

// Client is a struct that will interact with RocksDB
type Client struct {
	db           *boltdb.BoltDBWrapper
	ethClients   map[string]*ethereum.EthereumClient // key is node_url
	initialBlock int64
	blockStep    int64
	mutex        sync.Mutex // Add a mutex to control access to the critical sections
}

// NewClient initializes a new client with RocksDB
func NewClient(
	dbPath string,
	nodeUrls []string, // array of node URLs
	contractAddress string,
	topicToFilter string,
	initialBlock int64,
	blockStep int64,
) *Client {

	db := boltdb.NewBoltDB(dbPath)
	ethClients := make(map[string]*ethereum.EthereumClient)

	for _, nodeUrl := range nodeUrls {
		eth, err := ethereum.NewEthereumClient(
			nodeUrl,
			contractAddress,
			"contract/contractabi.json",
			topicToFilter,
		)
		if err != nil {
			log.Fatalf("Failed to create Ethereum client for node %s: %v", nodeUrl, err)
		}
		ethClients[nodeUrl] = eth
	}

	return &Client{
		db:           db,
		ethClients:   ethClients,
		initialBlock: initialBlock,
		blockStep:    blockStep,
	}
}

func (c *Client) Serve() {
	// Create a channel to manage work signals or steps
	stepChannel := make(chan int64, len(c.ethClients))

	// Initialize the starting block and step
	var initialBlock int64

	// Read the last block from BoltDB
	err, valueBlock := c.db.ReadFromDB("contract_handling", "last_block")
	if err != nil {
		log.Fatalf("Failed to read LAST BLOCK from BoltDB: %v", err)
	}
	if valueBlock == "" {
		initialBlock = c.initialBlock
		log.Printf("No last block found in DB. Starting from the initial block: %d", initialBlock)
	} else {
		initialBlock, err = strconv.ParseInt(valueBlock, 10, 64)
		if err != nil {
			log.Fatalf("Failed to convert last block to int64: %v", err)
		}
	}

	// Start workers for each Ethereum client
	for nodeUrl, ethClient := range c.ethClients {
		go func(nodeUrl string, ethClient *ethereum.EthereumClient) {
			for {
				// Send the current initial block and block step to the channel
				stepChannel <- initialBlock

				// Process blocks for this client
				c.processClient(nodeUrl, ethClient, stepChannel)
			}
		}(nodeUrl, ethClient)
	}

	// Optional: handle or log the results from the channel (if needed)
	for range stepChannel {
		// This loop can be used for logging or other operations if necessary
		log.Println("A worker has processed a step.")
	}
}

func (c *Client) processClient(nodeUrl string, ethClient *ethereum.EthereumClient, stepChannel <-chan int64) {
	for {
		// Wait for a signal from the step channel to start processing
		fromBlock := <-stepChannel

		// Unique keys for each client in the database
		lastBlockKey := "last_block"
		nextIndexKey := "next_index"

		// Lock the critical section where we read and update the database
		c.mutex.Lock()

		// Read last block from BoltDB
		err, valueBlock := c.db.ReadFromDB("contract_handling", lastBlockKey)
		if err != nil {
			log.Fatalf("[%s] Failed to read LAST BLOCK from BoltDB: %v", nodeUrl, err)
		}

		var lastProcessedBlock int64
		if valueBlock == "" {
			lastProcessedBlock = c.initialBlock
			log.Printf("[%s] No last block found in DB. Starting from the initial block: %d", nodeUrl, lastProcessedBlock)
		} else {
			lastProcessedBlock, err = strconv.ParseInt(valueBlock, 10, 64)
			if err != nil {
				log.Fatalf("[%s] Failed to convert last block to int64: %v", nodeUrl, err)
			}
		}

		// Read the last index from BoltDB
		err, valueIndex := c.db.ReadFromDB("contract_handling", nextIndexKey)
		if err != nil {
			log.Fatalf("[%s] Failed to read LAST INDEX from BoltDB: %v", nodeUrl, err)
		}

		var previousIndex int64
		if valueIndex == "" {
			previousIndex = 0
		} else {
			previousIndex, err = strconv.ParseInt(valueIndex, 10, 64)
			if err != nil {
				log.Fatalf("[%s] Failed to convert next_index to int64: %v", nodeUrl, err)
			}
		}

		toBlock := fromBlock + c.blockStep
		keyed, err := ethClient.KeyedLogs(fromBlock, toBlock)
		if err != nil {
			log.Fatalf("[%s] Failed to get keyed logs: %v", nodeUrl, err)
		}
		log.Printf("[%s] Keyed logs: %d", nodeUrl, len(keyed))

		// Unlock after reading and updating shared resources
		c.mutex.Unlock()

		// Process and write keyed logs
		for i, k := range keyed {
			keyedJson, err := json.Marshal(k)
			if err != nil {
				log.Fatalf("[%s] Error marshalling struct to JSON: %v", nodeUrl, err)
			}

			// Lock for critical section of writing the keyed logs
			c.mutex.Lock()

			// Write each keyed log into the database with the correct index
			indexKey := strconv.FormatInt(previousIndex+int64(i)+1, 10)
			if err := c.db.WriteToDB("keyed_logs", indexKey, string(keyedJson)); err != nil {
				log.Fatalf("[%s] Failed to write keyed log to BoltDB: %v", nodeUrl, err)
			}

			// Unlock after writing each keyed log
			c.mutex.Unlock()
		}

		// Lock again for writing the last block and index values
		c.mutex.Lock()

		// Save updated block and index references
		previousIndex += int64(len(keyed))
		if err := c.db.WriteToDB("contract_handling", nextIndexKey, fmt.Sprintf("%d", previousIndex)); err != nil {
			log.Fatalf("[%s] Failed to write next index to BoltDB: %v", nodeUrl, err)
		}
		if err := c.db.WriteToDB("contract_handling", lastBlockKey, fmt.Sprintf("%d", toBlock)); err != nil {
			log.Fatalf("[%s] Failed to write last block to BoltDB: %v", nodeUrl, err)
		}

		log.Printf("[%s] Saved last block: %d", nodeUrl, toBlock)
		log.Printf("[%s] Last Index: %d", nodeUrl, previousIndex)

		// Unlock after writing the last block and index
		c.mutex.Unlock()

		// Sleep before the next interval (customize as needed)
		time.Sleep(5 * time.Second) // Example delay for next round
	}
}

func (c *Client) LogKeyedInserts() {
	// GET LAST INDEX FROM BOLTDB
	err, valueIndex := c.db.ReadFromDB("contract_handling", "next_index")
	if err != nil {
		log.Fatalf("Failed to read NEXT INDEX from BoltDB: %v", err)
	}

	if valueIndex == "" {
		// No index in the database, nothing to show
		return
	}

	lastIndex, err := strconv.ParseInt(valueIndex, 10, 64)
	if err != nil {
		log.Fatalf("Failed to convert NEXT INDEX to int64: %v", err)
	}

	for i := int64(0); i < lastIndex; i++ {
		// Retrieve the value for the current index
		err, value := c.db.ReadFromDB("keyed_logs", strconv.FormatInt(i, 10))
		if err != nil {
			log.Fatalf("Failed to read from BoltDB: %v", err)
		}

		// Check if the value is empty
		if value == "" {
			log.Printf("Index:%d, WARNING: Empty value in DB.", i)
			continue
		}

		// Unmarshal the JSON into the struct
		var keyedLog domain.KeyedLog
		err = json.Unmarshal([]byte(value), &keyedLog)
		if err != nil {
			log.Printf("Index:%d, ERROR Failed to unmarshal JSON: %v", i, err)
		} else {
			log.Printf(
				"\tIndex: %d,\n\tRootData: %s,\n\tParentHash: %s,\n\tBlockTime: %d",
				i,
				keyedLog.RootData,
				keyedLog.ParentHash,
				keyedLog.BlockTime,
			)
		}
	}
}

// Close closes the BoltDB connection
func (c *Client) Close() {
	c.db.Close()
}
