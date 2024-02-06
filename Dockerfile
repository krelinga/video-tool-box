# syntax=docker/dockerfile:1

FROM krelinga/video-tool-box-base:buildx-latest AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
COPY pb/*.go ./pb/
RUN go mod download
RUN apt update
RUN apt install -y handbrake-cli

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /vtb .

ENV PWD=/
ENV VTB_NAS_MOUNT_DIR=/dev/null
ENV VTB_NAS_CANON_DIR=smb://dev/null

ENTRYPOINT ["/vtb", "--handbrake", "/usr/bin/HandBrakeCLI"]
