#!/bin/bash

## Build the go pacakge and run the binary within Linux containers

echo "[INFO] Building golang package for Linux binary"
# env GOOS=linux GOARCH=arm go build

cp -r /root/src/wordcounter /usr/local/go/src

cd /usr/local/go/src/wordcounter

go build

chmod +x ./wordcounter

./wordcounter