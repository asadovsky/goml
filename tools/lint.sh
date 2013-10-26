#!/bin/bash

set -e
set -u

cd $GOPATH

#find . -name '*.go' -print0 | xargs -0 gofmt -d
find . -name '*.go' -print0 | xargs -0 gofmt -w
