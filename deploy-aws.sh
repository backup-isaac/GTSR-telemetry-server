#!/bin/bash
rsync -r . ubuntu@solarracing.me:~/telemetry-server
ssh -t ubuntu@solarracing.me "sudo cp -r ~/telemetry-server /opt/telemetry-server; cd /opt/telemetry-server; sudo docker-compose up -d --force-recreate"