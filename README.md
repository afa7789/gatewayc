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
![alt text](image.png)
`make rund`
`make run`

# Example of code being Run

Just a screenshot of the code running from the scracth

![Example Image](resources/image.png)

# Comments

If code fails it's going to crash on fatal, so next steps would be to remove fatal, add retry mechanism and other things such as this to make sure the code can be left turned on. I'm considering it to be out of scope. So I'm not doing it.