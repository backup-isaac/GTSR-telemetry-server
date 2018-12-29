#!/bin/bash
docker volume create go-cache

if [[ $# -eq 0 ]] ; then
    docker run --rm=true -it -v ${PWD}:/go/src/rf-listener -v go-cache:/go/src -w /go/src/rf-listener --network="gtsr_default" golang:1.11.2 /bin/bash -c "go get go.bug.st/serial.v1 && go run listen.go"
else
    docker run --rm=true -it -v ${PWD}:/go/src/rf-listener -v go-cache:/go/src -w /go/src/rf-listener --device=$1 --network="gtsr_default" golang:1.11.2 /bin/bash -c "go get go.bug.st/serial.v1 && go run listen.go $1 $2"
fi



