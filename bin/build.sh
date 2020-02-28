#!/bin/bash -eu

source /build-common.sh

BINARY_NAME="alertmanager"
COMPILE_IN_DIRECTORY="cmd/alertmanager"

# aws has non-gofmt code..
GOFMT_TARGETS="cmd/ pkg/"

function packageLambdaFunction {
	cd rel/
	mv "${BINARY_NAME}_linux-amd64" "${BINARY_NAME}"
	rm -f alertmanager.zip
	# FIXME: zip is missing from image
	zip alertmanager.zip "${BINARY_NAME}"
}

standardBuildProcess

packageLambdaFunction
