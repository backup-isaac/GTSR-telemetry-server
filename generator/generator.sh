#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
docker run --rm -t -i --name generator -v $DIR:/app --network="telemetry-server_default" golang:1.11.2 go run /app/data_generator.go $1