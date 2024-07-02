#!/usr/bin/env bash

go build -v -o ./net/main ./net
go build -v -o ./uring/main ./uring

./net/main
./uring/main
