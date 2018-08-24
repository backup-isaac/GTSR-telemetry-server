#!/bin/bash
rsync -r . $1@mefsvs01.me.gatech.edu:/opt/telemetry-server
ssh -t $1@mefsvs01.me.gatech.edu "cd /opt/telemetry-server; sudo /usr/local/bin/docker-compose up -d --force-recreate --build"