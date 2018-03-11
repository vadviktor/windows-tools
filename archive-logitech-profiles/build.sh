#!/usr/bin/env bash

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -a -o archive-logitech-profiles.exe
