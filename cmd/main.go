package main

import (
	"fmt"
	"log"
	"os"

	"github.com/afa7789/gatewayc/internal/client"
	"github.com/joho/godotenv"
)

func main() {
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

	// Print Hello World and the ROCK_DB_HOST_PATH
	fmt.Printf("BOLT_DB_PATH: %s\n", boltDBpath)

	// Create a new client
	c := client.NewClient(boltDBpath, []string{})
	defer c.Close()

	// Serve the client
	c.Serve()
}
