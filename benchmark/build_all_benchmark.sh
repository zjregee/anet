#!/usr/bin/env bash

rm -rf output/
mkdir output
cp scripts/* output/

go build -v -o output/bencher ./bencher

mkdir -p output/net
go build -v -o output/net/bench ./net

mkdir -p output/gnet
go build -v -o output/gnet/bench ./gnet

mkdir -p output/netpoll
go build -v -o output/netpoll/bench ./netpoll

mkdir -p output/uring
go build -v -o output/uring/bench ./uring

mkdir -p output/anet
go build -v -o output/anet/bench ./anet
