#!/usr/bin/env bash

CURDIR=$(cd $(dirname $0); pwd)

if [ ! -d "$CURDIR/log" ]; then
    mkdir -p "$CURDIR/log"
fi

# export ANET_RUNTIME_LOG_FILE="$CURDIR/log/echo_server.log"

exec "$CURDIR/bin/echo_server"
