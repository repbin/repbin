#!/bin/sh
repclient --encrypt --repost --appdata -minDelay=1200 -maxDelay=7200 | repclient --encrypt --messageType=3 --recipientPubKey $1
