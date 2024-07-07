# syntax=docker/dockerfile:1

FROM krelinga/video-tool-box-base:buildx-latest AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /vtb .

ENV PWD=/
ENV VTB_NAS_MOUNT_DIR=/dev/null

ENTRYPOINT ["/vtb", "--handbrake", "/usr/bin/HandBrakeCLI"]
