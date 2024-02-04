#! /usr/bin/bash

export PATH="${PATH}:$HOME/go/bin"

rm -f *.pb.go
protoc \
    -I=. \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    *.proto
