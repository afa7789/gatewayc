package main

import (
	"log"
	"os"
	"strconv"
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

	initialBlock := os.Getenv("INITIAL_BLOCK")
	if initialBlock == "" {
		log.Fatal("INITIAL_BLOCK is not set in the environment")
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
	log.Printf("BOLT_DB_PATH: %s\n", boltDBpath)
	log.Printf("CONTRACT_ADDRESS: %s\n", contractAddress)
	log.Printf("TOPIC_TO_FILTER: %s\n", topicToFilter)
	log.Printf("NODE_URL: %s\n", nodeUrl)
	//
	// Create a new client
	initialBlockInt, err := strconv.ParseInt(initialBlock, 10, 64)
	if err != nil {
		log.Fatalf("Failed to convert initialBlock to int64: %v", err)
	}

	c := client.NewClient(
		boltDBpath,
		nodeUrl,
		contractAddress,
		topicToFilter,
		initialBlockInt,
	)
	defer c.Close()

	// Serve the client
	c.Serve()
}
