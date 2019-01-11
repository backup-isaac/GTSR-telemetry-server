#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
docker build -t generator $DIR
docker run --rm -t -i --name generator --network="telemetry-server_default" generator go run data_generator.go $1