#!/bin/bash

pwd=$(pwd)
cd $(dirname $(realpath $0))

export LOG_MODE="dev"
export LOG_LEVEL="d"
# export DB_ADDRESS="10.20.22.165:30100"
export DB_ADDRESS="localhost:9000"
export CONF="e30K" # {}
# export GOMAXPROCS=2
# go run . $@
CGO_ENABLED=0 go run . 
#taskset -c 0-3 
cd $pwd
