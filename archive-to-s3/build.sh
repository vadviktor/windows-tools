#!/usr/bin/env bash

if [ -z "$1" ]; then
    echo "No output filename given"
    exit 1
fi

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -a -o $1
