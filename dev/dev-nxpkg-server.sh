#!/bin/sh

set -ex

# Build a Nxpkg server docker image to run for development purposes. Note
# that this image is not exactly identical to the published nxpkg/server
# images, as those include Nxpkg's proprietary code behind paywalls.
time cmd/server/pre-build.sh
IMAGE=nxpkg/server:$USER-dev VERSION=$USER-dev time cmd/server/build.sh

IMAGE=nxpkg/server:$USER-dev ${BASH_SOURCE%/*}/run-server-image.sh
