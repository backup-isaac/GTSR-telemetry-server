#!/bin/bash
rsync -r . ubuntu@18.221.138.32:~/telemetry-server
ssh -t ubuntu@18.221.138.32 "sudo cp -r ~/telemetry-server /opt; cd /opt/telemetry-server; sudo docker-compose up -d --force-recreate"