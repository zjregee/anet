#!/usr/bin/env bash

rm -rf output/
mkdir output
cp scripts/* output/

mkdir -p output/net
go build -v -o output/net/bench_serial ./net/serial
go build -v -o output/net/bench_concurrent ./net/concurrent

mkdir -p output/netpoll
go build -v -o output/netpoll/bench_serial ./netpoll/serial
go build -v -o output/netpoll/bench_concurrent ./netpoll/concurrent

mkdir -p output/uring
go build -v -o output/uring/bench_serial ./uring/serial
go build -v -o output/uring/bench_concurrent ./uring/concurrent
