#!/bin/bash
docker exec -it server /bin/sh -c 'kill $(ps -aux | grep generator | awk "{print \$2}")'