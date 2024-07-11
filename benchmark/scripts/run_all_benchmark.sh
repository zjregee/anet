#!/usr/bin/env bash

./net/bench -c=12 -m=1000 -n=100 -len=256
./net/bench -c=24 -m=1000 -n=100 -len=256
./net/bench -c=36 -m=1000 -n=100 -len=256
./net/bench -c=12 -m=1000 -n=100 -len=1024
./net/bench -c=24 -m=1000 -n=100 -len=1024
./net/bench -c=36 -m=1000 -n=100 -len=1024

./netpoll/bench -c=12 -m=1000 -n=100 -len=256
./netpoll/bench -c=24 -m=1000 -n=100 -len=256
./netpoll/bench -c=36 -m=1000 -n=100 -len=256
./netpoll/bench -c=12 -m=1000 -n=100 -len=1024
./netpoll/bench -c=24 -m=1000 -n=100 -len=1024
./netpoll/bench -c=36 -m=1000 -n=100 -len=1024

./uring/bench -c=12 -m=1000 -n=100 -len=256
./uring/bench -c=24 -m=1000 -n=100 -len=256
./uring/bench -c=36 -m=1000 -n=100 -len=256
./uring/bench -c=12 -m=1000 -n=100 -len=1024
./uring/bench -c=24 -m=1000 -n=100 -len=1024
./uring/bench -c=36 -m=1000 -n=100 -len=1024

./anet/bench -c=12 -m=1000 -n=100 -len=256
./anet/bench -c=24 -m=1000 -n=100 -len=256
./anet/bench -c=36 -m=1000 -n=100 -len=256
./anet/bench -c=12 -m=1000 -n=100 -len=1024
./anet/bench -c=24 -m=1000 -n=100 -len=1024
./anet/bench -c=36 -m=1000 -n=100 -len=1024
