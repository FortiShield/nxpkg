#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd $(dirname "${BASH_SOURCE[0]}")/../..
set -ex

GOBIN=$PWD/vendor/.bin go install ./vendor/github.com/nxpkg/godockerize

# Additional images passed in here when this script is called externally by our
# enterprise build scripts.
additional_images=${@:-github.com/nxpkg/nxpkg/cmd/frontend}

# Overridable server package path for when this script is called externally by
# our enterprise build scripts.
server_pkg=${SERVER_PKG:-github.com/nxpkg/nxpkg/cmd/server}

./vendor/.bin/godockerize build --base 'alpine:3.8' -t ${IMAGE} --go-build-flags="-ldflags" --go-build-flags="-X github.com/nxpkg/nxpkg/pkg/version.version=${VERSION}" --env VERSION=${VERSION} \
    $server_pkg \
    github.com/nxpkg/nxpkg/cmd/github-proxy \
    github.com/nxpkg/nxpkg/cmd/gitserver \
    github.com/nxpkg/nxpkg/cmd/query-runner \
    github.com/nxpkg/nxpkg/cmd/symbols \
    github.com/nxpkg/nxpkg/cmd/repo-updater \
    github.com/nxpkg/nxpkg/cmd/searcher \
    github.com/nxpkg/nxpkg/cmd/indexer \
    github.com/nxpkg/nxpkg/vendor/github.com/google/zoekt/cmd/zoekt-archive-index \
    github.com/nxpkg/nxpkg/vendor/github.com/google/zoekt/cmd/zoekt-nxpkg-indexserver \
    github.com/nxpkg/nxpkg/vendor/github.com/google/zoekt/cmd/zoekt-webserver \
    github.com/nxpkg/nxpkg/cmd/lsp-proxy $additional_images
