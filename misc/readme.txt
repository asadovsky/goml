export GOPATH=${HOME}/dev/goml

${GOPATH}/tools/lint.sh

go test goml/core goml/algo
go test goml/algo -cpu=4
