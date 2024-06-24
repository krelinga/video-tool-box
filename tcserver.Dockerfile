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
COPY tcserver/hb/*.go ./tcserver/hb/
COPY tcserver/transcoder/*.go ./tcserver/transcoder/
COPY tcserver/transcoder/related/*.go ./tcserver/transcoder/related/
COPY tcserver/transcoder/show/*.go ./tcserver/transcoder/show/

RUN CGO_ENABLED=0 GOOS=linux go build -o /tcserver ./tcserver/

ENV VTB_TCSERVER_PORT=25000
ENV VTB_TCSERVER_IN_PATH_PREFIX="smb://truenas/media"
ENV VTB_TCSERVER_OUT_PATH_PREFIX="/videos"
ENV VTB_TCSERVER_PROFILE="mkv_h265_2160p60_fast"
ENV VTB_TCSERVER_FILE_WORKERS=1
ENV VTB_TCSERVER_MAX_QUEUED_FILES=10000
ENV VTB_TCSERVER_SHOW_WORKERS=1
ENV VTB_TCSERVER_MAX_QUEUED_SHOWS=10
ENV VTB_TCSERVER_SPREAD_WORKERS=1
ENV VTB_TCSERVER_MAX_QUEUED_SPREADS=10

ENTRYPOINT ["/usr/bin/bash", "-c", "/tcserver"]

EXPOSE 25000
