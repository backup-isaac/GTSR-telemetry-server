#!/bin/bash
if [[ $# -eq 0 ]] ; then
    echo 'Usage: ./proxy-jenkins.sh <your GT username>'
    exit 1
fi
echo "Setting up ssh tunnels to localhost 8088"
ssh $1@mefsvs01.me.gatech.edu -L 8088:mefsvs01.me.gatech.edu:8080 -N

