# syntax=docker/dockerfile:1

FROM krelinga/video-tool-box-base:buildx-latest AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
# TODO: we'll want handbrake eventually, but no need to have it for hello world.
# RUN apt update
# RUN apt install -y handbrake-cli

# TODO: it would be nice to have a better way to recursively copy _all_
# the source files to their correct directories...
COPY *.go ./
COPY pb/*.go ./pb/
COPY tcserver/*.go ./tcserver/

RUN CGO_ENABLED=0 GOOS=linux go build -o /tcserver ./tcserver/

ENV VTB_TCSERVER_PORT=25000

ENTRYPOINT ["/usr/bin/bash", "-c", "/tcserver"]

EXPOSE 25000
