# Basic go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Go 1.13.1 commands
GO113CMD=~/go/bin/go1.13.1
GO113BUILD=$(GOCMD) build
GO113CLEAN=$(GOCMD) clean
GO113TEST=$(GOCMD) test
GO113GET=$(GOCMD) get

# Binary names
BINARY_NAME=battleship
BINARY_UNIX=$(BINARY_NAME)_unix

export PATH := $(PATH):~/go/bin

all: download build
build: 
		$(GO113BUILD) -o $(BINARY_NAME) -v ./cmd/web 
clean: 
		$(GO113CLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
deps:
		$(GO113GET) golang.org/x/xerrors
		$(GO113GET) github.com/justinas/alice
		$(GO113GET) github.com/mattn/go-sqlite3

download:
		$(GOGET) golang.org/dl/go1.13.1
		$(GO113CMD) download
run:
		$(GO113BUILD) -o $(BINARY_NAME) -v ./...
		./$(BINARY_NAME)
test: 
		$(GO113TEST) -v ./...


# Cross compilation
build-linux:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
