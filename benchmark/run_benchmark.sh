#!/usr/bin/env bash

go build -v -o ./net/bench_serial ./net/serial
go build -v -o ./net/bench_concurrent ./net/concurrent
go build -v -o ./netpoll/bench_serial ./netpoll/serial
go build -v -o ./netpoll/bench_concurrent ./netpoll/concurrent
go build -v -o ./uring/bench_serial ./uring/serial
go build -v -o ./uring/bench_concurrent ./uring/concurrent

./net/bench_serial
./net/bench_concurrent
./netpoll/bench_serial
./netpoll/bench_concurrent
./uring/bench_serial
./uring/bench_concurrent
