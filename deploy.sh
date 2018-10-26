#!/bin/bash
if [[ $# -eq 0 ]] ; then
    echo 'Usage: ./deploy.sh <your GT username>'
    exit 1
fi
echo "Now copying to ~/telemetry-server"
rsync -r . $1@mefsvs01.me.gatech.edu:~/telemetry-server
echo "Now copying to /opt/telemetry-server and restarting (This takes a while)"
ssh -t $1@mefsvs01.me.gatech.edu "sudo systemctl stop telemetry-server; sudo cp -r ~/telemetry-server /opt; cd /opt/telemetry-server; sudo systemctl start telemetry-server"

