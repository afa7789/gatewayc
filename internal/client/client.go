// client/client.go
package client

import (
	"log"

	"github.com/afa7789/gatewayc/internal/boltdb"
	"github.com/afa7789/gatewayc/internal/ethereum"
)

// Client is a struct that will interact with RocksDB
type Client struct {
	db  *boltdb.BoltDBWrapper
	eth *ethereum.EthereumClient
}

// NewClient initializes a new client with RocksDB
func NewClient(
	dbPath string,
	nodeUrl string,
	contractAddress string,
	topicToFilter string,

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

	return &Client{db: db, eth: eth}
}

func (c *Client) Serve() {
	err := c.db.WriteToDB("bucket", "hello_world", "Hello World!")
	if err != nil {
		log.Fatalf("Failed to write 'Hello World!' to RocksDB: %v", err)
	}

	err, value := c.db.ReadFromDB("bucket", "hello_world")
	if err != nil {
		log.Fatalf("Failed to read from RocksDB: %v", err)
	}
	log.Printf("Successfully read '%s' from the key 'hello_world'.", value)

	c.eth.FetchLogs(
		6525866,
		6525867,
	)
}

// Close closes the RocksDB connection
func (c *Client) Close() {
	c.db.Close()
}
