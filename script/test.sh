#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct

name="localhost"
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-s -w" -o plugin/fetch-$name test/plugin/$name.go

if [ "$1" = "report" ]; then
  go test -cover -covermode=atomic -coverprofile=coverage.txt -parallel 2 -race -v ./...
else
  list="$(go list ./... | grep -v foo)"
  old=$IFS IFS=$'\n'
  for item in $list; do
    go test -cover -covermode=atomic -parallel 2 -race -v "$item"
  done
  IFS=$old
fi
