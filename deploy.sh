#!/bin/bash
rsync -r . $1@mefsvs01.me.gatech.edu:~/telemetry-server
ssh -t $1@mefsvs01.me.gatech.edu "sudo cp -r ~/telemetry-server /opt/telemetry-server; cd /opt/telemetry-server; sudo /usr/local/bin/docker-compose up -d"