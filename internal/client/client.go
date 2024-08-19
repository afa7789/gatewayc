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
	// Channel to send block ranges to workers
	blockRangeChannel := make(chan struct{ from, to int64 })

	// Initialize the starting block
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

	// Create workers
	for nodeUrl, ethClient := range c.ethClients {
		go func(nodeUrl string, ethClient *ethereum.EthereumClient) {
			c.processClient(nodeUrl, ethClient, blockRangeChannel)
		}(nodeUrl, ethClient)
	}

	// Distribute block ranges
	go func() {
		for {
			toBlock := initialBlock + c.blockStep
			blockRangeChannel <- struct{ from, to int64 }{from: initialBlock, to: toBlock}
			initialBlock = toBlock
			time.Sleep(5 * time.Second) // Adjust this as needed to control the distribution frequency
		}
	}()

	// Keep the Serve function running
	select {} // This will block indefinitely to keep the Serve function alive
}

func (c *Client) processClient(nodeUrl string, ethClient *ethereum.EthereumClient, blockRangeChannel <-chan struct{ from, to int64 }) {
	for blockRange := range blockRangeChannel {
		fromBlock, toBlock := blockRange.from, blockRange.to

		// Retrieve logs
		keyed, err := ethClient.KeyedLogs(fromBlock, toBlock)
		if err != nil {
			log.Fatalf("[%s] Failed to get keyed logs: %v", nodeUrl, err)
		}

		// generate a name, with a mask just 4 last digits of node_url
		nodeUrlMask := nodeUrl[len(nodeUrl)-4:]

		log.Printf("[%s] Block range: %d-%d", nodeUrlMask, fromBlock, toBlock)
		log.Printf("[%s] Keyed logs: %d", nodeUrlMask, len(keyed))

		// Lock for critical section of reading and writing the lastIndex
		c.mutex.Lock()

		// Read the last index from BoltDB
		err, valueIndex := c.db.ReadFromDB("contract_handling", "next_index")
		if err != nil {
			log.Fatalf("[%s] Failed to read NEXT INDEX from BoltDB: %v", nodeUrl, err)
		}

		var lastIndex int64
		if valueIndex == "" {
			lastIndex = 0
		} else {
			lastIndex, err = strconv.ParseInt(valueIndex, 10, 64)
			if err != nil {
				log.Fatalf("[%s] Failed to convert NEXT INDEX to int64: %v", nodeUrl, err)
			}
		}

		// Process and write keyed logs
		for i, k := range keyed {
			keyedJson, err := json.Marshal(k)
			if err != nil {
				log.Fatalf("[%s] Error marshalling struct to JSON: %v", nodeUrl, err)
			}

			// Write keyed log with the correct index
			indexKey := strconv.FormatInt(lastIndex+int64(i)+1, 10)
			if err := c.db.WriteToDB("keyed_logs", indexKey, string(keyedJson)); err != nil {
				log.Fatalf("[%s] Failed to write keyed log to BoltDB: %v", nodeUrl, err)
			}
		}

		// Update last index and last block in BoltDB
		lastIndex += int64(len(keyed))
		if err := c.db.WriteToDB("contract_handling", "next_index", fmt.Sprintf("%d", lastIndex)); err != nil {
			log.Fatalf("[%s] Failed to write next index to BoltDB: %v", nodeUrl, err)
		}
		if err := c.db.WriteToDB("contract_handling", "last_block", fmt.Sprintf("%d", toBlock)); err != nil {
			log.Fatalf("[%s] Failed to write last block to BoltDB: %v", nodeUrl, err)
		}

		// Unlock after writing
		c.mutex.Unlock()
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
