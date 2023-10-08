#!/bin/bash

set -ex

export POD_NAME="dev"
export NAMESPACE="dev"
export ELEMENT_NAME="dev"

export LOG_MODE="mjson"

# go run . abc
go run . bash -c 'while [ 1 ]; do echo abc; sleep 1; done'
