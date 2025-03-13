#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <version-code>"
  exit 1
fi

VERSION_CODE=$1
RELEASE_DIR="releases/$VERSION_CODE"
mkdir -p "$RELEASE_DIR"

GOOS=windows GOARCH=amd64 go build -o "$RELEASE_DIR/starbit-client-windows.exe" ./client
GOOS=darwin GOARCH=amd64 go build -o "$RELEASE_DIR/starbit-client-macos" ./client
GOOS=linux GOARCH=amd64 go build -o "$RELEASE_DIR/starbit-client-linux" ./client


GOOS=linux GOARCH=amd64 go build -o "$RELEASE_DIR/starbit-server-linux" ./server

echo "Build completed. Executables are located in $RELEASE_DIR" 