#!/usr/bin/env bash

CURDIR=$(cd $(dirname $0); pwd)
export ANET_RUNTIME_LOG_FILE="$CURDIR/log/http_server.log"
exec "$CURDIR/bin/http_server"
