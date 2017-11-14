#!/bin/bash

set -x

SRC="$( cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

GEOIP_DIR=$SRC/assets/geoip

TMPFILE=$GEOIP_DIR/GeoLite2-Country.mmdb.gz

mkdir -p $GEOIP_DIR

# download database
if [ ! -f $TMPFILE ]; then
  wget -O $TMPFILE http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz
fi

# package
go-bindata \
  -ignore .go$ \
  -ignore .gitignore \
  -pkg geoip \
  -prefix $GEOIP_DIR \
  -o $GEOIP_DIR/geoip.go \
  $GEOIP_DIR/...
