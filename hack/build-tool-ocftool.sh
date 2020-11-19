#!/bin/bash
set -euE

ARCHs="amd64"
OSes="linux darwin windows"

for ARCH in $ARCHs; do
  for OS in $OSes; do
    binary="bin/ocftool-$OS-$ARCH"

    GOOS=$OS GOARCH=$ARCH go build -o "$binary" cmd/ocftool/main.go
  done
done
