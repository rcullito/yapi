#!/bin/bash

# Init reqs
clear

# Init env
SCRIPT_DIR=`pwd -P`
export PATH=$PATH:$GOPATH/bin:$SCRIPT_DIR

echo ==============================
echo GO Environment
echo ==============================
go env
echo ------------------------------
