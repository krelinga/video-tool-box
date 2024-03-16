# syntax=docker/dockerfile:1

FROM golang:1.21



WORKDIR /app

# TODO: it would be nice to have a better way to recursively copy _all_
# the source files to their correct directories...
COPY go.mod go.sum ./
COPY pb/*.go ./pb/
COPY mkvinfoserver/*.go ./mkvinfoserver/
COPY mkvinfoserver_deps.sh ./

RUN ./mkvinfoserver_deps.sh

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /mkvinfoserver ./mkvinfoserver/

ENV VTB_MKVINFOSERVER_PORT=25001

ENTRYPOINT ["/usr/bin/bash", "-c", "/mkvinfoserver"]

EXPOSE 25001
