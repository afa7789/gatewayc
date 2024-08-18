package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/afa7789/gatewayc/internal/client"
	"github.com/joho/godotenv"
)

func main() {
	log.Default().Print("Starting the GatewayC client")
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get the ROCK_DB_HOST_PATH environment variable
	boltDBpath := os.Getenv("BOLT_DB_PATH")
	if boltDBpath == "" {
		log.Fatal("BOLT_DB_PATH is not set in the environment")
	}

	contractAddress := os.Getenv("CONTRACT_ADDRESS")
	if contractAddress == "" {
		log.Fatal("CONTRACT_ADDRESS is not set in the environment")
	}

	topicToFilter := os.Getenv("TOPIC_TO_FILTER")
	if topicToFilter == "" {
		log.Fatal("TOPIC_TO_FILTER is not set in the environment")
	}

	// read first line from file in nodeList
	// to get a node provider RPC url
	nodeListPath := os.Getenv("NODES_LIST")
	if nodeListPath == "" {
		log.Fatal("NODES_LIST is not set in the environment")
	}
	nodeListFile, err := os.ReadFile(nodeListPath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	nodeList := strings.Split(string(nodeListFile), "\n")
	nodeUrl := nodeList[0]

	// Print Hello World and the ROCK_DB_HOST_PATH
	fmt.Printf("BOLT_DB_PATH: %s\n", boltDBpath)

	// Create a new client
	c := client.NewClient(
		boltDBpath,
		nodeUrl,
		contractAddress,
		topicToFilter,
	)
	defer c.Close()

	// Serve the client
	c.Serve()
}
