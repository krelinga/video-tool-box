# syntax=docker/dockerfile:1

FROM krelinga/video-tool-box-base:buildx-latest AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# TODO: it would be nice to have a better way to recursively copy _all_
# the source files to their correct directories...
COPY *.go ./
COPY pb/*.go ./pb/
COPY tcserver/*.go ./tcserver/

RUN CGO_ENABLED=0 GOOS=linux go build -o /tcserver ./tcserver/

ENV VTB_TCSERVER_PORT=25000
ENV VTB_TCSERVER_IN_PATH_PREFIX="smb://truenas/media"
ENV VTB_TCSERVER_OUT_PATH_PREFIX="/videos"
ENV VTB_TCSERVER_PROFILE="mkv_h265_2160p60_very_fast"

ENTRYPOINT ["/usr/bin/bash", "-c", "/tcserver"]

EXPOSE 25000
