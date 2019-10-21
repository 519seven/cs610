#!/bin/bash

[[ -f .install_env ]] && { printf "Sourcing .install_env\n"; . .install_env; } || { printf ".install_env not found. Exiting.\n"; exit -1; }

GOVERSION=$go_version
GOPROGRAM=$go_program
if [[ ! -d ./tls ]]; then
  mkdir -p tls
fi  
cd tls
if [ ! -f cert.pem ] || [ ! -f key.pem ]; then
  printf "Finding your home directory\n"
  homedir=$(getent passwd "$USER" | cut -d: -f6)
  printf "Using $homedir as your home directory\n"
  printf "Searching for generate_cert.go (this may take a while)\n"
  gencert=$(find $homedir -type d -name '*go*' -exec find {} -name generate_cert.go -print \;)
  printf "Generating RSA key pair using $GOPROGRAM and $gencert...\n"
  $GOPROGRAM run $gencert --rsa-bits=2048 --host=localhost
  printf "Storing private key in key.pem file\n"
  printf "Generating self-signed TLS certificate for localhost and storing in cert.pem file\n"
else
  printf "Certs already exist!\n"
fi

