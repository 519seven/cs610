#!/bin/bash

CURDIR=$(shell pwd)

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
# Grab the old values before altering them
export OLDGOROOT=$GOROOT && printf "Saved your old GOROOT ($GOROOT) in OLDGOROOT in case you needed that\n"
unset GOROOT && printf "unsetting GOROOT\n"
export OLDGOPATH=$GOPATH && printf "Saved your old GOPATH ($GOPATH) in OLDGOPATH in case you needed that\n"
export GOPATH=$(CURDIR)/go && printf "Set GOPATH to $GOPATH\n"
export PATH=$PATH:$GOPATH/bin && printf "Updating PATH to $PATH\n"
rc=$(go get golang.org/dl/go1.13)
if [[ $rc -eq 0 ]]; then
  printf "Getting go1.13\n"
  go1.13 download
else
  printf "Failure to get go1.13\nI am unsure how to continue :(\n"
  exit 1
fi


