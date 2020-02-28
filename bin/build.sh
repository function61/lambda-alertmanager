#!/bin/bash -eu

source /build-common.sh

BINARY_NAME="alertmanager"
COMPILE_IN_DIRECTORY="cmd/alertmanager"

# aws has non-gofmt code..
GOFMT_TARGETS="cmd/ pkg/"

# TODO: one deployerspec is done, we can stop overriding this from base image
function packageLambdaFunction {
	cd rel/
	cp "${BINARY_NAME}_linux-amd64" "${BINARY_NAME}_lambda"
	rm -f lambdafunc.zip
	zip lambdafunc.zip "${BINARY_NAME}_lambda"
	rm "${BINARY_NAME}_lambda"
}

standardBuildProcess

packageLambdaFunction
