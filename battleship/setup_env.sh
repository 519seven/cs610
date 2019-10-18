#!/bin/bash

CURDIR=$(pwd)

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
printf "Checking for go1.13\n"
GO=$(( which go1.13 ) 2>&1)
if [[ $? -ne 0 ]]; then
  printf "go1.13 not found.  Setting up...\n"
  # Grab the old values before altering them
  export OLDGOROOT=$GOROOT && printf "Saved your old GOROOT ($GOROOT) in OLDGOROOT in case you needed that\n"
  unset GOROOT && printf "unsetting GOROOT\n"
  export OLDGOPATH=$GOPATH && printf "Saved your old GOPATH ($GOPATH) in OLDGOPATH in case you needed that\n"
  export GOPATH=$CURDIR/go && printf "Set GOPATH to $GOPATH\n"
  export PATH=$PATH:$GOPATH/bin && printf "Updating PATH to $PATH\n"
  rc="$(go get golang.org/dl/go1.13 2>&1 > /dev/null)"
  if [[ $rc == *"permission denied"* ]]; then
    printf "Getting go1.13\n"
    $GO download
  else
    [[ ! -z $err_msg ]] && printf "Error encountered: $err_msg"
    printf "Failure to get go1.13\nI am unsure how to continue :(\n"
    exit 1
  fi
else
  printf "go1.13 is already installed\nPassing control back to make setup\n"
fi

