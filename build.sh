#!/bin/bash
# This script builds the project using the specified build tool.

rm -rf ./dist
GOOS=linux GOARCH=amd64 go build -o blacklight_linux_amd64 .
GOOS=linux GOARCH=arm64 go build -o blacklight_linux_arm64 .
GOOS=windows GOARCH=amd64 go build -o blacklight_windows_amd64.exe .
GOOS=windows GOARCH=arm64 go build -o blacklight_windows_arm64.exe .
GOOS=darwin GOARCH=amd64 go build -o blacklight_darwin_amd64 .
GOOS=darwin GOARCH=arm64 go build -o blacklight_darwin_arm64 .

mkdir -p ./dist
mv blacklight_* ./dist/