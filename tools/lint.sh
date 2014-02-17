#!/bin/bash

set -e
set -u

cd $GOPATH

gofmt -w .
