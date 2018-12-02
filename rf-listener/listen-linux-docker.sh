#!/bin/bash

docker build -t rf-listener .

if [[ $# -eq 0 ]] ; then
    docker run --rm=true --name=rf-listener -it --network="telemetry-server_default" rf-listener go run listen.go
else
    docker run --rm=true --name=rf-listener -it --device=$1 --network="telemetry-server_default" rf-listener go run listen.go $1 $2
fi



