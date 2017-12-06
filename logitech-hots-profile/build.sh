#!/usr/bin/env bash

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -o /media/sf_vmshared/logitech-hots-profile.exe
