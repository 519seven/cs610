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
BINARY_NAME=./battleship
DB_FILE=./battleship.db
GO_ARTIFACTS=._*
TLS_CERTS=./tls/*.pem

all: setup certs build

test:
		env

build: 
		@set -e; ./build_project.sh	&& { 										\
			echo "---------------- Project build succeeded! -----------------";	\
			echo "---------- Run ./battleship (-h for more info)) -----------";	\
			echo "------------------ Your battle awaits you! ----------------"; \
			exit 0; 															\
		 } || {																	\
			echo "Project build failed. Check output.";							\
			exit 1;																\
		 }
		
certs:
		@set -e; ./gen_cert.sh && {												\
			echo "Certificates are in place! Next is build.";					\
		 } || {																	\
		 	echo "Cert we not generated. Check output."; 						\
			exit 1;																\
		 }

clean: 
		@rm -f $(BINARY_NAME)
		@rm -f $(ENV_FILE)
		@echo "Removing $(BINARY_NAME) and $(ENV_FILE)..."
		@rm -f $(GO_ARTIFACTS)
		@echo "Removing hidden files that cause build failures..."
		@rm -f $(TLS_CERTS)
		@echo "Removing TLS certificates..."
		@rm -f $(DB_FILE)
		@echo "Removing DB file..."
		@echo "Done!"

setup:
		@./setup_env.sh && {													\
			echo "Set up complete!";											\
		 } || { 																\
		 	echo "Setup exiting!";												\
			exit 1;																\
		 }
