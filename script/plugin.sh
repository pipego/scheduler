#!/bin/bash

list="plugin-fetch,plugin-filter,plugin-score"

if [ ! -d bin ]; then
  mkdir bin
fi

old=$IFS IFS=$','
for item in $list; do
  URL=$(curl -L -s https://api.github.com/repos/pipego/$item/releases/latest | grep -o -E "https://(.*)${item}_(.*)_linux_amd64.tar.gz")
  curl -L -s "$URL" | tar xvz -C bin
done
IFS=$old
