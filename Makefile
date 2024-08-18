#!/bin/bash
include .env

# run the node
run: ## calling the cmd to run the node.
	@echo "\033[2m→ Running the project...\033[0m"
	@docker run -it my-go-app /bin/bash

# docker
docker:
	@docker build -t my-go-app .
