#!/bin/sh -eu

rm -f canary.zip && cd src/ && zip ../canary.zip index.js utils.js
