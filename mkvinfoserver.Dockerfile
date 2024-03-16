# syntax=docker/dockerfile:1

FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# TODO: it would be nice to have a better way to recursively copy _all_
# the source files to their correct directories...
COPY pb/*.go ./pb/
COPY tcserver/*.go ./mkvinfoserver/

RUN CGO_ENABLED=0 GOOS=linux go build -o /mkvinfoserver ./mkvinfoserver/

ENV VTB_MKVINFOSERVER_PORT=25001

ENTRYPOINT ["/usr/bin/bash", "-c", "/mkvinfoserver"]

EXPOSE 25001
