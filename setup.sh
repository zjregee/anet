#!/usr/bin/env bash

git clone -b liburing-2.7 https://github.com/axboe/liburing.git
cd liburing
./configure --cc=gcc --cxx=g++
make -j$(nproc)
sudo make install
cd ../
rm -rf liburing

go install golang.org/x/tools/cmd/goimports@latest
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.3
