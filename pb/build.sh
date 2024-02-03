#! /usr/bin/bash

export PATH="${PATH}:$HOME/go/bin"
protoc -I=. --go_out=. --go_opt=paths=source_relative *.proto
