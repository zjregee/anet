#!/usr/bin/env bash

./net/bench -c=4 -m=2000000 -len=1024
./net/bench -c=12 -m=2000000 -len=1024
./net/bench -c=64 -m=2000000 -len=1024
./net/bench -c=128 -m=2000000 -len=1024
./net/bench -c=512 -m=2000000 -len=1024
./net/bench -c=1024 -m=2000000 -len=1024

./netpoll/bench -c=4 -m=2000000 -len=1024
./netpoll/bench -c=12 -m=2000000 -len=1024
./netpoll/bench -c=64 -m=2000000 -len=1024
./netpoll/bench -c=128 -m=2000000 -len=1024
./netpoll/bench -c=512 -m=2000000 -len=1024
./netpoll/bench -c=1024 -m=2000000 -len=1024

./uring/bench -c=4 -m=2000000 -len=1024
./uring/bench -c=12 -m=2000000 -len=1024
./uring/bench -c=64 -m=2000000 -len=1024
./uring/bench -c=128 -m=2000000 -len=1024
./uring/bench -c=512 -m=2000000 -len=1024
./uring/bench -c=1024 -m=2000000 -len=1024

./anet/bench -c=4 -m=2000000 -len=1024
./anet/bench -c=12 -m=2000000 -len=1024
./anet/bench -c=64 -m=2000000 -len=1024
./anet/bench -c=128 -m=2000000 -len=1024
./anet/bench -c=512 -m=2000000 -len=1024
./anet/bench -c=1024 -m=2000000 -len=1024
