#! /usr/bin/bash
#
if ! /usr/bin/docker buildx ls | egrep -q "^multiarch" ; then
    /usr/bin/docker buildx create \
        --name multiarch \
        --bootstrap
fi

/usr/bin/docker buildx build --builder=multiarch --push \
    --platform linux/amd64,linux/arm64 \
    -f tcserver.Dockerfile \
    --tag krelinga/video-tool-box-tcserver:buildx-latest .
