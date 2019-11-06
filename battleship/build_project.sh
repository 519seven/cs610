#!/bin/bash

if [[ -f .install_env ]]; then
    . .install_env;
    ${GO} build -o battleship -v ./cmd/web
    exit 0;
else
    printf ".install_env not found. Try running `make`.\n";
    exit 1;
fi