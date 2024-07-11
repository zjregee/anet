#!/usr/bin/env bash

git clone https://github.com/axboe/liburing.git
cd liburing
./configure --cc=gcc --cxx=g++
make -j$(nproc)
make install
cd ../
rm -rf liburing
