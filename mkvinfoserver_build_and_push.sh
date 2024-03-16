#! /usr/bin/bash

/usr/bin/docker buildx build --builder=myBuilder --push \
    --platform linux/amd64,linux/arm64 \
    -f mkvinfoserver.Dockerfile \
    --tag krelinga/video-tool-box-mkvinfoserver:buildx-latest .
