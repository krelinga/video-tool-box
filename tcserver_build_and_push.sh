#! /usr/bin/bash

/usr/bin/docker buildx build --builder=myBuilder --push \
    --platform linux/amd64,linux/arm64 \
    -f tcserver.Dockerfile \
    --tag krelinga/video-tool-box-tcserver:buildx-latest .
