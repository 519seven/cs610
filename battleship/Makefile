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

include .install_env
export $(shell sed 's/=.*//' .install_env)

# Binary names
BINARY_NAME=battleship
BINARY_UNIX=$(BINARY_NAME)_unix

all: setup certs build

test:
		env

build: 
		${GO} build -o $(BINARY_NAME) -v ./cmd/web
		$(info 	Your battle awaits you!  Run ./battleship (-h for more info))
		$(info  ------------------ Remember to use HTTPS ------------------)
		
certs:
		set -e; ./gen_cert.sh

clean: 
		$(GO113CLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
deps:
		$(GO113GET) golang.org/x/xerrors
		$(GO113GET) github.com/justinas/alice
		$(GO113GET) github.com/mattn/go-sqlite3
		$(GO113GET) github.com/golangcollege/sessions@v1

run:
		${GO} -o $(BINARY_NAME) -v ./...
		./$(BINARY_NAME)
setup:
		set -e; ./setup_env.sh

build-linux:
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
