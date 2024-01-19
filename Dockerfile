# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN apt update
RUN apt install -y handbrake-cli

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /vtb .

ENTRYPOINT ["/vtb", "--handbrake", "/usr/bin/HandBrakeCLI"]
