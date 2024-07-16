#!/usr/bin/env bash

gcc -shared -o libsum.so -fPIC sum.c
go test -bench=.
rm libsum.so
