#!/bin/bash
docker exec -it $(docker ps | grep telemetry-server_server | awk '{print $NF}') /bin/sh -c 'kill $(ps -aux | grep generator | awk "{print \$2}")'