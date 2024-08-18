#!/bin/bash
include .env

# run the node
run: ## calling the cmd to run the node.
	@echo "\033[2m→ Running the project...\033[0m"
	@go run cmd/main.go

# docker
docker:
	@docker build -t my-go-app .
	@docker run --rm my-go-app

# build
build:
	@go build -o /app/main ./cmd/main.go
