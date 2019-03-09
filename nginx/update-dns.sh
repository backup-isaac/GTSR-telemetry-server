#!/usr/bin/env bash

domain-resolver.sh /etc/nginx/allowed-domain.list > /etc/nginx/allowed-ips-from-domains.conf
service nginx reload > /dev/null 2>&1
