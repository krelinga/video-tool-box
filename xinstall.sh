#! /usr/bin/bash

export GOOS=darwin
export GOARCH=arm64

go build -o /host/macbin/vtb .
