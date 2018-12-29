#!/bin/bash
# Only to be used by Jenkins -- do not call directly
set -e
for fn in $(find . -name go.mod); do
    dn=$(dirname $fn)
    cd $dn
    go build -o main
    rm main
    cd ..
done