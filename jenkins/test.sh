#!/bin/bash
# Only to be used by Jenkins -- do not call directly
set -e
for fn in $(find . -name go.mod); do
    dn=$(dirname $fn)
    cd $dn
    if [[ $(gofmt -l -s .) ]]; then
        exit 1
    fi
    go test ./...
    cd ..
done