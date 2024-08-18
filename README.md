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

you willl need docker to execute everything

## EXECUTING

### DOCKER
with docker ( so you don't have to install anything ):

`make docker`
`make run`

if the docker is already built and you have no code changes, just use:

`make rund`
`make run`