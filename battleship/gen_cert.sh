#!/bin/bash

[[ -f .install_env ]] && { printf "Sourcing .install_env\n"; . .install_env; } || { printf ".install_env not found. Exiting.\n"; exit -1; }

if [[ ! -d ./tls ]]; then
  mkdir -p tls
fi  
cd tls
if [[ -z $GOBASE ]]; then
  printf "GOBASE is empty, unable to continue...\n"
  exit 0
fi
if [ ! -f cert.pem ] || [ ! -f key.pem ]; then
  printf "Finding your home directory\n"
  homedir=$(getent passwd "$USER" | cut -d: -f6)
  printf "Using $homedir as your home directory\n"
  printf "Searching for generate_cert.go (this may take a while)\n"
  gencert=$(find $homedir -type f -name generate_cert.go -print -exec find {} -name $GOBASE \;|head -1)
  printf "gencert: $gencert\nGO: $GO\n"
  printf "Generating RSA key pair...\n"
  $GO run $gencert --rsa-bits=2048 --host=localhost
  if [ $? -eq 1 ]; then
    printf "Error generating certificates! Exiting...\n"
    exit 1
  fi
  printf "Storing private key in key.pem file\n"
  printf "Generating self-signed TLS certificate for localhost and storing in cert.pem file\n"
else
  printf "Certs already exist!\n"
fi
printf -- "------------------ Remember to use HTTPS ------------------\n"
exit 0
