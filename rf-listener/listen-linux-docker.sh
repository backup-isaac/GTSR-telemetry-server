#!/bin/bash

if [[ $# -eq 0 ]] ; then
    docker run --rm=true -it -v ${PWD}:/go/src/rf-listener -v ${PWD}/go/src:/go/src -w /go/src/rf-listener --network="telemetry-server_default" golang:1.11.2 /bin/bash -c "go get go.bug.st/serial.v1 && go run listen.go"
else
    docker run --rm=true -it -v ${PWD}:/go/src/rf-listener -v ${PWD}/go/src:/go/src -w /go/src/rf-listener --device=$1 --network="telemetry-server_default" golang:1.11.2 /bin/bash -c "go get go.bug.st/serial.v1 && go run listen.go $1 $2"
fi



