#!/bin/bash

set -e

CURDIR=$(pwd)
INSTALLGO="N"
USER_GOROOT=/usr/local

function get_answer {
  read -p "Would you like to set it up for your user (no sudo required) [y/n]? " yesno
  case "$yesno" in
    y|Y)
      printf "$GOBASE will be installed\n"
      INSTALLGO=true
      return 0
      ;;
    n|N)
      printf "Not installing $GOBASE\n"
      INSTALLGO=false
      exit 1
      ;;
    *)
      printf %s\\n "Enter y or n only"
      return 1
      ;;
  esac
}

function go_get_var {
  # Try to find and set our go vars
  printf "Checking for go1.13 (this may take several seconds)\n"
  GO=$(which go)
  if [ $? -eq 0 ]; then
    # Success
    GOVER=$($GO version)
    if [[ $GOVER == *"1.13"* ]]; then
      # Found version 1.13
      printf "$GOVER was found, able to continue...\n"
      GOBASE=$(basename $GO 2>/dev/null)
    else
      # Look in user's home directory for the downloader 
      GO=$(find ~ -type f -name 'go1.13' -o -name 'go1.13.1' | grep bin | head -1)
      if [ $? -eq 0 ]; then
        GOBASE=$(basename $GO 2>/dev/null)
        if ! $GO version 2>/dev/null; then
          # go was found but it isn't installed
          GO=""
        fi
      else
        printf "go is not version 1.13\n"
        exit 1
      fi
    fi
  else
    printf "go was not found on your system. Please install version 1.13\n"
    exit 1
  fi
}

function go_found {
  # Look for an existing go version
  if [[ $GO == *"go1.13"* || $GO == *"go1.13.1"* || $GOVER == *"1.13"* ]]; then
    save_env
    printf "$GOVER was found.  You're good to go :)\n"
    printf "Passing control back to setup...\n"
    exit 0
  fi
  printf "Go $GOBASE was not found.\n"
  until get_answer; do : ; done
}

function go_save_vars {
  # Grab the old values before altering them
  export OLDGOROOT=$GOROOT && printf "Saved your old GOROOT ($GOROOT) in OLDGOROOT in case you needed that\n"
  unset GOROOT && printf "unsetting GOROOT\n"
  export OLDGOPATH=$GOPATH && printf "Saved your old GOPATH ($GOPATH) in OLDGOPATH in case you needed that\n"
  export GOPATH=$CURDIR/go && printf "Set GOPATH to $GOPATH\n"
  export PATH=$PATH:$GOPATH/bin && printf "Updating PATH to $PATH\n"
}

function go_get_go {
  # Download go1.13
  rc="$(go get golang.org/dl/go1.13 2>&1 > /dev/null)"
  if [[ $rc == *"permission denied"* ]]; then
    printf "Failure to get go1.13\nI am unsure how to continue :(\n"
    exit 1
  else
    printf "Getting go1.13\n"
    rc="$($GOBASE download 2>&1 >/dev/null)"
    if [[ $rc == *"already downloaded"* ]]; then
      printf "$GOBASE is already downloaded. You're good to go :)"
    fi
    save_env
    exit 0
  fi
}

function save_env {
  echo -e "export GO=$GO\nexport GOBASE=$GOBASE" > .install_env
}
# -----------------------------------------------------------------------------

go_get_var || { printf "Error in go_get_var. Exiting...\n"; exit 1; }
if [[ $GOVER != *"1.13"* ]]; then
  go_found || { printf "Error in go_found. Exiting...\n"; exit 2; }
  go_save_vars || { printf "Error in go_save_vars. Exiting...\n"; exit 3; }
  go_get_var || { printf "Error in go_get_var. Exiting...\n"; exit 4; }
  go_get_go || { printf "Error in go_get_go. Exiting...\n"; exit 5; }
else
  printf "You have the correct go version\n"
fi