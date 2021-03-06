#!/bin/bash
# Run this script before committing changes

go get golang.org/x/lint/golint

for fn in $(find . -name go.mod); do
    dn=$(dirname $fn)
    cd $dn
    gofmt -l -w -s .
    go test ./...
    go mod tidy
    $(go env GOPATH)/bin/golint ./...
    cd ..
done