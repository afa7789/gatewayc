// client/client.go
package client

import (
	"log"

	"github.com/afa7789/gatewayc/internal/boltdb"
)

// Client is a struct that will interact with RocksDB
type Client struct {
	db *boltdb.BoltDBWrapper
}

// NewClient initializes a new client with RocksDB
func NewClient(
	dbPath string,
	nodesList []string,
) *Client {
	db := boltdb.NewBoltDB(dbPath)
	return &Client{db: db}
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
}

// Close closes the RocksDB connection
func (c *Client) Close() {
	c.db.Close()
}
