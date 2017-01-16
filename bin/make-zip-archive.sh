#!/bin/sh -eu

date_formatted=`date +%Y-%m-%d`

filename_canary="alertmanager-canary-$date_formatted.zip"
filename_alertmanager="alertmanager-$date_formatted.zip"

(rm -f "$filename_canary" && cd alertmanager-canary/ && zip "../$filename_canary" index.js utils.js)

(rm -f "$filename_alertmanager" && cd alertmanager && zip "../$filename_alertmanager" index.js)
