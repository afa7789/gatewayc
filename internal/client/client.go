// client/client.go
package client

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/afa7789/gatewayc/internal/boltdb"
	"github.com/afa7789/gatewayc/internal/domain"
	"github.com/afa7789/gatewayc/internal/ethereum"
)

// Client is a struct that will interact with RocksDB
type Client struct {
	db           *boltdb.BoltDBWrapper
	eth          *ethereum.EthereumClient
	initialBlock int64
	blockStep    int64
}

// NewClient initializes a new client with RocksDB
func NewClient(
	dbPath string,
	nodeUrl string,
	contractAddress string,
	topicToFilter string,
	initialBlock int64,
	blockStep int64,
) *Client {

	db := boltdb.NewBoltDB(dbPath)
	eth, err := ethereum.NewEthereumClient(
		nodeUrl,
		contractAddress,
		"contract/contractabi.json",
		topicToFilter,
	)
	if err != nil {
		log.Fatalf("Failed to create Ethereum client: %v", err)
	}

	return &Client{db: db, eth: eth, initialBlock: initialBlock, blockStep: blockStep}
}

func (c *Client) Serve() {

	// testing the db
	err := c.db.WriteToDB("bucket", "hello_world", "Hello World!")
	if err != nil {
		log.Fatalf("Failed to write 'Hello World!' to BoltDB: %v", err)
	}

	err, value := c.db.ReadFromDB("bucket", "hello_world")
	if err != nil {
		log.Fatalf("Failed to read from BoltDB: %v", err)
	}
	log.Printf("Successfully read '%s' from the key 'hello_world'.", value)

	// PUT THIS IN AN INTERVAL
	for {
		// GET LAST BLOCK FROM BOLTDB
		err, valueBlock := c.db.ReadFromDB("contract_handling", "last_block")
		if err != nil {
			log.Fatalf("Failed to read LAST BLOCK from BoltDB: %v", err)
		}

		var fromBlock int64

		// Check if valueBlock is empty or nil
		if valueBlock == "" || len(valueBlock) == 0 {
			// If there is no block in the db, start from the initial block
			fromBlock = c.initialBlock
			log.Printf("No last block found in DB. Starting from the initial block: %d", fromBlock)
		} else {
			// Try to parse the block value if it exists
			fromBlock, err = strconv.ParseInt(valueBlock, 10, 64)
			if err != nil {
				log.Fatalf("Failed to convert last block to int64: valueBlock=%v, error: %v", valueBlock, err)
			}
		}

		// GET LAST INDEX FROM BOLTDB
		err, valueIndex := c.db.ReadFromDB("contract_handling", "next_index")
		if err != nil {
			log.Fatalf("Failed to read LAST INDEX from BoltDB: %v", err)
		}

		var previousIndex int64

		// Check if valueIndex is empty or nil before parsing
		if valueIndex == "" || len(valueIndex) == 0 {
			// If there is no index in the db, start from the initial index (0)
			previousIndex = 0
			log.Printf("No last index found in DB. Starting from index: %d", previousIndex)
		} else {
			// If valueIndex is not empty, attempt to parse it
			previousIndex, err = strconv.ParseInt(valueIndex, 10, 64)
			if err != nil {
				log.Fatalf("Failed to convert next_index to int64: valueIndex=%v, error: %v", valueIndex, err)
			}
		}

		toBlock := fromBlock + c.blockStep
		keyed, err := c.eth.KeyedLogs(
			fromBlock,
			toBlock,
		)
		if err != nil {
			log.Fatalf("Failed to get keyed logs: %v", err)
		}
		log.Printf("Keyed logs: %v", len(keyed))

		for i, k := range keyed {
			// marshal keyed
			// Marshal the struct to JSON
			keyedJson, err := json.Marshal(k)
			if err != nil {
				log.Fatalf("Error marshalling struct to JSON: %v", err)
			}

			// Convert index to string
			indexKey := strconv.FormatInt(int64(i)+previousIndex+1, 10)

			// store at BoltDB
			if err := c.db.WriteToDB("keyed_logs", indexKey, string(keyedJson)); err != nil {
				// Must come up with a better way to handle error here...
				// probably also need a retry logic
				log.Fatalf("Failed to write keyed log to BoltDB: %v", err)
			}
		}

		// save references, for next interval
		// these values must be saved on the DB, in case the application crash and you must run it again.
		previousIndex += int64(len(keyed))
		if err := c.db.WriteToDB("contract_handling", "next_index", fmt.Sprintf("%d", previousIndex)); err != nil {
			log.Fatalf("Failed to write last block to BoltDB: %v", err)
		}
		if err := c.db.WriteToDB("contract_handling", "last_block", fmt.Sprintf("%d", toBlock)); err != nil {
			log.Fatalf("Failed to write last block to BoltDB: %v", err)
		}
		log.Printf("Saved last block: %d", toBlock)
		log.Printf("Last Index: %d", previousIndex)
		// END interval
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
