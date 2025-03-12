#! /bin/bash

MAX_PLAYERS=4
if [ $# -gt 0 ]; then
    MAX_PLAYERS=$1
fi

echo "Starting server with maxPlayers=$MAX_PLAYERS"

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/service.proto
go run ./server -maxPlayers=$MAX_PLAYERS