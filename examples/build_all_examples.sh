#!/usr/bin/env bash

rm -rf output/
mkdir output
mkdir -p output/bin
mkdir -p output/log
cp scripts/* output/

go build -v -o output/bin/tcp_server ./tcp_server
