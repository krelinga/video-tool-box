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

ENTRYPOINT ["/usr/bin/bash", "-c", "/vtb", "--handbrake", "/usr/bin/HandBrakeCLI"]
