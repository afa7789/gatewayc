// client/client.go
package client

import (
	"encoding/json"
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
		if valueBlock == "" {
			// if there is no block in the db, start from the initial block
			fromBlock = c.initialBlock
		} else {
			fromBlock, err = strconv.ParseInt(valueBlock, 10, 64)
			if err != nil {
				log.Fatalf("Failed to convert initialBlock to int64: %v", err)
			}
		}

		// GET LAST BLOCK INDEX FROM BOLTDB
		err, valueIndex := c.db.ReadFromDB("contract_handling", "last_index")
		if err != nil {
			log.Fatalf("Failed to read LAST BLOCK from BoltDB: %v", err)
		}
		previousIndex, err := strconv.ParseInt(valueIndex, 10, 64)
		if err != nil {
			log.Fatalf("Failed to convert initialBlock to int64: %v", err)
		} else if valueIndex == "" {
			// if there is no block in the db, start from the initial block
			previousIndex = 0
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

			// store at BoltDB
			if err := c.db.WriteToDB("keyed_logs", string(int64(i)+previousIndex), string(keyedJson)); err != nil {
				// Must come up with a better way to handle error here...
				// probably also need a retry logic
				log.Fatalf("Failed to write keyed log to BoltDB: %v", err)
			}
		}

		// save references, for next interval
		// these values must be saved on the DB, in case the application crash and you must run it again.
		previousIndex += int64(len(keyed) + 1)
		if err := c.db.WriteToDB("contract_handling", "last_index", string(previousIndex)); err != nil {
			log.Fatalf("Failed to write last block to BoltDB: %v", err)
		}
		if err := c.db.WriteToDB("contract_handling", "last_block", string(toBlock)); err != nil {
			log.Fatalf("Failed to write last block to BoltDB: %v", err)
		}
		// END interval
	}

}

// This would be a way to retrieve the values at the db.
func (c *Client) LogKeyedInserts() {
	err, valueIndex := c.db.ReadFromDB("contract_handling", "last_index")
	if err != nil {
		log.Fatalf("Failed to read LAST BLOCK from BoltDB: %v", err)
	}

	if valueIndex == "" {
		// if there is no index in db
		// nothing to show
		return
	}

	lastIndex, err := strconv.ParseInt(valueIndex, 10, 64)
	if err != nil {
		log.Fatalf("Failed to convert initialBlock to int64: %v", err)
	}

	for i := int64(0); i < lastIndex; i++ {
		err, value := c.db.ReadFromDB(
			"keyed_logs",
			strconv.FormatInt(i, 10),
		)
		if err != nil {
			log.Fatalf("Failed to read from BoltDB: %v", err)
		}

		// string to bytes
		var keyedLog domain.KeyedLog
		err = json.Unmarshal([]byte(value), &keyedLog)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		log.Printf(
			"\tIndex: %d,\n\tRootData: %s,\n\tParentHash: %s,\n\tBlockTime: %d",
			i,
			keyedLog.RootData,
			keyedLog.ParentHash,
			keyedLog.BlockTime,
		)
	}

}

// Close closes the BoltDB connection
func (c *Client) Close() {
	c.db.Close()
}
