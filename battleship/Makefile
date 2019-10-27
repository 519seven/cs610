#!make

# Basic go commands
#GOCMD=go
#GOBUILD=$(GOCMD) build
#GOCLEAN=$(GOCMD) clean
#GOTEST=$(GOCMD) test
#GOGET=$(GOCMD) get

# Go 1.13.1 commands
#GO113CMD=go/bin/go
#GO113BUILD=$(GO113CMD) build
#GO113CLEAN=$(GO113CMD) clean
#GO113TEST=$(GO113CMD) test
#GO113GET=$(GO113CMD) get

# Environment setup
ENV_FILE=.install_env

# Binary names
BINARY_NAME=battleship


all: setup certs build

test:
		env

build: 
		set -e; ./build_project.sh
		
certs:
		set -e; ./gen_cert.sh

clean: 
		rm -f $(BINARY_NAME)
		rm -f $(ENV_FILE)

setup:
		set -e; ./setup_env.sh
