#!/bin/bash

SRC="$( cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

LOCS="en id es"

TEMPLATE_DIR=$SRC/assets/templates
LOCALE_DIR=$SRC/assets/locales
POTFILE=$LOCALE_DIR/messages.pot

set -ev

pushd $TEMPLATE_DIR &> /dev/null

# extract messages
go-xgettext \
  --package-name=gophers-web \
  --keyword=T \
  --keyword-plural=N \
  --sort-output \
  *.go > $POTFILE

perl -pi -e 's/build\/templates\///g' $POTFILE

popd &> /dev/null

# generate / merge the extracted messages for each locale
for LOC in $LOCS; do
  POOUT=$LOCALE_DIR/$LOC.po
  #MOOUT=$LOCALE_DIR/$LOC.mo

  if [ ! -f $POOUT ]; then
    # create initial messages.po if it doesn't exist
    msginit --locale=$LOC --no-wrap -i $POTFILE -o $POOUT
  else
    # merge messages otherwise
    msgmerge --update --no-wrap $POOUT $POTFILE
  fi

  # generate mo
  #msgfmt -o $MOOUT $POOUT
done
