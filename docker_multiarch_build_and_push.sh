#! /usr/bin/bash

/usr/bin/docker buildx build --push \
    --platform linux/amd64,linux/arm64 \
    --tag krelinga/video-tool-box:buildx-latest .