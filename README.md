![Theme Image](resources/banner.png)

# Indexer

just a simple implementation of a indexer, that will index the topic X of contract Y in chain sepolia

`for second challenge check branch "second"`

## Pre-Requisites && Installing

GO LANG version: `go1.22.1'`

GNU Make

```SH
#MAC OS
brew install make

# REDHAT FEDORA CENTOS
sudo dnf install make
# or for older versions
sudo yum install make

#DEBIAN
sudo apt update
sudo apt install make
```

rocksdb, you have to follow install from here, if not using docker: https://github.com/facebook/rocksdb

if you have have docker, just do the docker running option.

## EXECUTING

### DOCKER
with docker ( so you don't have to install anything ):

`make docker`

### RUNNING IT
`make run`

### BINARY
#### build binary

`make bin`

#### run binary

`make serve`