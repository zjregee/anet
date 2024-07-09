#!/usr/bin/env bash

./net/bench_serial
./net/bench_concurrent
./netpoll/bench_serial
./netpoll/bench_concurrent
./uring/bench_serial
./uring/bench_concurrent
