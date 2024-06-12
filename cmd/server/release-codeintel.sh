#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Illegal number of parameters. Please read ./dev/server/README.md"
    exit -1
fi

lang="$1"
version="$2"
outlang="$lang"

if [ "$(echo "$outlang" | tail -c 7)" = "skinny" ]; then
    outlang="${outlang%-skinny}"
fi

echo "nxpkg/xlang-$lang:$version => nxpkg/codeintel-$outlang:$version (and latest)"
echo -n 'Continue? [y/N] '
read shouldProceed
if [ "$shouldProceed" != "y" ] && [ "$shouldProceed" != "Y" ]; then
    echo Aborting
    exit 1
fi

set -ex

docker pull "nxpkg/xlang-$lang:$version"
docker tag "nxpkg/xlang-$lang:$version" "nxpkg/codeintel-$outlang:$version"
docker tag "nxpkg/xlang-$lang:$version" "nxpkg/codeintel-$outlang:latest"
docker push "nxpkg/codeintel-$outlang:$version"
docker push "nxpkg/codeintel-$outlang:latest"
