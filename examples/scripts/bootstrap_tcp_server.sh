#!/usr/bin/env bash

CURDIR=$(cd $(dirname $0); pwd)
export ANET_RUNTIME_LOG_FILE="$CURDIR/log/tcp_server.log"
exec "$CURDIR/bin/tcp_server"
