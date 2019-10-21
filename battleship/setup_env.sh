#!/bin/bash

CURDIR=$(pwd)
INSTALLGO="N"
USER_GOROOT=/usr/local

# Try # 1
#mkdir -p ~/go/go1.13.1
#export GOROOT=~/go/go1.13.1
#export GOPATH=~/go/src
#export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
#go get golang.org/dl/go1.13.1

# Try #2
#export GOROOT=/usr
#export GOPATH=~/go
#export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Try #3
printf "Checking for go1.13 (this may take several seconds)\n"
GO113=$(find ~ -type f -name 'go1.13' -o -name 'go1.13.1' | grep bin | head -1)
GOVER=$($GO113 version | awk '{print $3}')
if [[ $GOVER == "go1.13" || $GOVER == "go1.13.1" ]]; then
  echo "GO=$GOVER" > .install_env
  printf "Go version 1.13 was found.  You're good to go :)\n"
  printf "Passing control back to setup...\n"
  exit 0
fi
printf "Go 1.13 was not found.  Attempting to set it up...\n"
GO=$(( which go ) 2>&1)
# Grab the old values before altering them
export OLDGOROOT=$GOROOT && printf "Saved your old GOROOT ($GOROOT) in OLDGOROOT in case you needed that\n"
unset GOROOT && printf "unsetting GOROOT\n"
export OLDGOPATH=$GOPATH && printf "Saved your old GOPATH ($GOPATH) in OLDGOPATH in case you needed that\n"
export GOPATH=$CURDIR/go && printf "Set GOPATH to $GOPATH\n"
export PATH=$PATH:$GOPATH/bin && printf "Updating PATH to $PATH\n"
rc="$(go get golang.org/dl/go1.13 2>&1 > /dev/null)"
if [[ $rc == *"permission denied"* ]]; then
  printf "Failure to get go1.13\nI am unsure how to continue :(\n"
  exit 1
else
  printf "Getting go1.13\n"
  rc="$($GOVER download 2>&1 >/dev/null)"
  if [[ $rc == *"already downloaded"* ]]; then
    printf "Go1.13 is already downloaded. You're good to go :)"
  fi
  echo "GO=go1.13" > .install_env
  exit 0
fi