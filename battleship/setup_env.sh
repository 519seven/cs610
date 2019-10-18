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
printf "Checking for go1.13\n"
# Check for go
which go 2>/dev/null
gorc=$?
if [[ $gorc -ne 0 ]]; then
  printf "I am unable to find a go installation on this server, period\n"
  printf "Do you want me to install go?\n[y/N]: "
  read INSTALLGO
  INSTALLGO=$(echo $INSTALLGO | tr [a-z] [A-Z])
  if [[ $INSTALLGO == 'Y' ]]; then
    which apt
    if [[ $? -eq 0 ]]; then
      printf "Attempting to updating your system.\n"
      printf "If this is undesirable, select N at the confirmation dialog.\n"
      sudo apt-get update
      sudo apt-get upgrade
    else
      which yum
      if [[ $? -eq 0 ]]; then
        printf "Attempting to update your system.\n"
        printf "If this is undesirable, select N at the confirmation dialog.\n"
        sudo yum update
      fi
    fi
    USER_GOROOT=/usr/local
    printf "Installing go\n"
    printf "Where is your GOROOT? ($USER_GOROOT is typical default): "
    read USER_GOROOT
    if [[ $USER_GOROOT != /usr/local && $USER_GOROOT != '' ]]; then
      printf "Sorry, you want to put go somewhere other than a typical default location\n"
      printf "This script isn't intelligent enough to handle this customization\n"
      printf "Exiting...\n"
      exit -1
    elif [[ $USER_GOROOT == /usr/local || $USER_GOROOT == '' ]]; then
      which wget
      if [[ $? -ne 0 ]]; then
        printf "Please install wget before proceeding\n"
        exit -1
      fi
      wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz
      sudo tar -xvf go1.13.linux-amd64.tar.gz
      sudo mv go /usr/local
      export GOROOT=$USER_GOROOT/go
      export GOPATH=$(pwd)
      export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
      echo "GOROOT=$GOROOT" > .install_env
      echo "GOPATH=$GOPATH" >> .install_env
      echo "PATH=$GOPATH/bin:$GOROOT/bin:$PATH" >> .install_env
      go_version=$(go version | awk '{print $3}')
      echo "go_program=$(which go)" >> .install_env
      echo "go_version=$go_version" >> .install_env
      which go
      gorc=$?
      if [[ $gorc -ne 0 ]]; then
        printf "Well, I tried.  Something isn't right.  Go installation can't be found.  Exiting...\n"
        exit -1
      fi
    else
      printf "Please set up your go environment.  I'm unsure what to do with the information received.\n"
      exit -1
    fi
  else
    printf "You've said you don't want me to install go so I'm unable to continue :(\n"
    exit 1
  fi
else
  printf "A go installation was found\n"
  echo "GOROOT=$USER_GOROOT/go" > .install_env
  echo "GOPATH=$(pwd)" >> .install_env
  echo "PATH=$GOPATH/bin:$GOROOT/bin:$PATH" >> .install_env
  go_version=$(go version | awk '{print $3}')
  echo "go_program=$(which go)" >> .install_env
  echo "go_version=$go_version" >> .install_env
fi

if [[ $gorc -eq 0 ]]; then
  printf "Checking for go1.13\n"
  go_version=$(go version | awk '{print $3}')
  if [[ $go_version == "go1.13" ]]; then
    printf "go1.13 is already installed\nSwitching control back to make setup\n"
    go_version=$(go version | awk '{print $3}')
    go_program=$(which go)
    echo "go_version=$go_version" >> .install_env
    exit 0
  else
    printf "go1.13 not found.  Setting up...\n"
    rc="$(go get golang.org/dl/go1.13 2>&1 > /dev/null)"
    if [[ $rc == *"permission denied"* ]]; then
      printf "Getting go1.13\n"
      go1.13 download
    else
      [[ ! -z $err_msg ]] && printf "Error encountered: $err_msg"
      printf "Failure to get go1.13\nI am unsure how to continue :(\n"
      exit 1
    fi
  fi
fi
