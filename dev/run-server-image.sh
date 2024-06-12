#!/bin/bash
# This runs a published or local server image.

DATA=/tmp/nxpkg

case $CLEAN in

    "true")
        clean=y
        ;;
    "false")
        clean=n
        ;;
    *)
        echo -n "Do you want to delete $DATA and start clean? [Y/n] "
        read clean
        ;;
esac

if [ "$clean" != "n" ] && [ "$clean" != "N" ]; then
    echo "deleting $DATA"
    rm -rf $DATA
fi

IMAGE=${IMAGE:-nxpkg/server:insiders}
echo "pulling docker image ${IMAGE}"
docker pull $IMAGE

echo "starting server..."
docker run "$@" \
 --publish 7080:7080 --rm \
 -e SRC_LOG_LEVEL=dbug \
 -e DEBUG=t \
 --volume $DATA/config:/etc/nxpkg \
 --volume $DATA/data:/var/opt/nxpkg \
 -v /var/run/docker.sock:/var/run/docker.sock \
 $IMAGE
