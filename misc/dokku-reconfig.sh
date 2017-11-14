#!/bin/bash

REMOTE=$(git config remote.dokku.url)

HOST=$(echo $REMOTE|awk -F: '{print $1}')
APP=$(echo $REMOTE|awk -F: '{print $2}')

ssh -t $HOST apps:create $APP

dokku-push-config APP_CONFIG env/config

ssh -t $HOST config:set $APP ENV=production
