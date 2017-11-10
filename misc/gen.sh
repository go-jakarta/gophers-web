#!/bin/bash

#set -ex

VER=1.9.2

if [ -z "$BUILD_DIR" ]; then
  BUILD_DIR=$(pwd)
fi

CACHE_DIR=$BUILD_DIR/.cache

mkdir -p $CACHE_DIR

if [ ! -d $CACHE_DIR/go-$VER ]; then
  pushd $CACHE_DIR &> /dev/null

  if [ ! -f go$VER.tar.gz ]; then
    wget -q -O go$VER.tar.gz "https://storage.googleapis.com/golang/go$VER.linux-amd64.tar.gz"
  fi

  rm -rf go-$VER

  tar -zxf go$VER.tar.gz

  mv go go-$VER

  popd &> /dev/null
fi

export GOROOT=$CACHE_DIR/go-$VER
export PATH=$CACHE_DIR/go-$VER/bin:$PATH

go generate

go build
