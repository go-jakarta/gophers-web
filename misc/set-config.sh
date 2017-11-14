#!/bin/bash

# this will put a file into memory

CONFIG="$1"
if [ ! -f "$CONFIG" ]; then
  CONFIG=./env/config
fi

export APP_CONFIG=$(base64 -w 0 "$CONFIG")
