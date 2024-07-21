#!/usr/bin/env bash

./bencher -c=128 -m=10000000 -len=1024 -port=:8000 -name=net
./bencher -c=128 -m=10000000 -len=1024 -port=:8000 -name=netpoll
./bencher -c=128 -m=10000000 -len=1024 -port=:8000 -name=uring
./bencher -c=128 -m=10000000 -len=1024 -port=:8000 -name=anet
