#!/bin/bash -eu

source /build-common.sh

standardBuildProcess "backend"

zip ../rel/alertmanager-canary.zip *.js
