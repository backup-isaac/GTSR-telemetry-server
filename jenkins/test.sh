#!/bin/bash
# Only to be used by Jenkins -- do not call directly
set -e
for fn in $(find . -name go.mod); do
    dn=$(dirname $fn)
    cd $dn
    malformatted=$(gofmt -l -s .)
    if [[ "$malformatted" ]]; then
        echo "error: package $dn has formatting error(s) in files: $malformatted" >&2
        exit 1
    fi
    go test ./...
    cd ..
done
