#!/usr/bin/env bash

rm -rf output/
mkdir -p output/bin
cp scripts/* output/

go build -v -o output/bin/echo_server ./echo_server

cd output/
./bootstrap_echo_server.sh
