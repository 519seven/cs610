#!/bin/bash

[[ -f .install_env ]] && { printf "Sourcing .install_env\n"; . .install_env; } || { printf ".install_env not found. Exiting.\n"; exit -1; }

${GO} build -o battleship -v ./cmd/web
printf "Your battle awaits you!  Run ./battleship (-h for more info))\n"
printf -- "------------------ Remember to use HTTPS ------------------\n"