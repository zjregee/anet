#!/usr/bin/env bash

git clone -b liburing-2.7 https://github.com/axboe/liburing.git
cd liburing
./configure --cc=gcc --cxx=g++
make -j$(nproc)
sudo make install
cd ../
rm -rf liburing
