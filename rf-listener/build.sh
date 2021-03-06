#!/bin/bash 
docker volume create go-cache

docker run --rm=true -it -v ${PWD}:/go/src/rf-listener -v go-cache:/go/src -w /go/src/rf-listener golang:1.13 /bin/bash -c \
    "GOOS=linux GOARCH=amd64 go build -o bin/listen listen.go \
        && GOOS=darwin GOARCH=amd64 go build -o bin/listen.app listen.go \
        && GOOS=windows GOARCH=amd64 go build -o bin/listen.exe listen.go"
