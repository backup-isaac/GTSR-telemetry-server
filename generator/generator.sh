#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
docker run --rm -t -i --name generator -v $DIR:/app --network="telemetry-server" golang:1.13 go run /app/data_generator.go /app/route_receiver.go $1 $2
