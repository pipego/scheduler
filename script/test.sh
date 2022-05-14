#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct

filepath="test/plugin"
for item in $(ls $filepath); do
  if [ -d "$filepath"/"$item" ]; then
    for name in "$filepath"/"$item"/*.go; do
      buf=${name%.go}
      name=${buf##*/}
      CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-s -w" -o plugin/"$item"-"$name" "$filepath"/"$item"/"$name".go
    done
  fi
done

if [ "$1" = "report" ]; then
  go test -cover -covermode=atomic -coverprofile=coverage.txt -parallel 2 -race -v ./...
else
  list="$(go list ./... | grep -v test)"
  old=$IFS IFS=$'\n'
  for item in $list; do
    go test -cover -covermode=atomic -parallel 2 -race -v "$item"
  done
  IFS=$old
fi
