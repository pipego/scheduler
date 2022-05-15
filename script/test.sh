#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct

filepath="test/plugin"
for item in $(ls $filepath); do
  if [ -d "$filepath"/"$item" ]; then
    for dname in $(ls "$filepath"/"$item"); do
      if [ -d "$filepath"/"$item"/"$dname" ]; then
        for name in "$filepath"/"$item"/"$dname"/*.go; do
          buf=${name%.go}
          name=${buf##*/}
          # go tool dist list
          CGO_ENABLED=0 GOARCH=$(go env GOARCH) GOOS=$(go env GOOS) go build -ldflags "-s -w" -o plugin/"$item"-"$name" "$filepath"/"$item"/"$dname"/"$name".go
        done
      fi
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
