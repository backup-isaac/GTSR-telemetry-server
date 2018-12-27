#!/bin/bash
# Run this script before committing changes

go get golang.org/x/lint/golint

for fn in $(find . -name go.mod); do
    dn=$(dirname $fn)
    cd $dn
    go mod download
    go mod tidy
    go fmt ./...
    go test ./...
    $(go env GOPATH)/bin/golint ./...
    cd ..
done