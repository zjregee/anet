#!/usr/bin/env bash

go build -v -o ./net/bench_serial ./net/serial
go build -v -o ./net/bench_concurrent ./net/concurrent
go build -v -o ./uring/bench_serial ./uring/serial

./net/bench_serial
./net/bench_concurrent
./uring/bench_serial
