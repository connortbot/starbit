# Quick Start

## Setup
```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate proto code (run this after changing proto files)
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/service.proto
```

## Run
```bash
# Terminal 1: Start server
go run ./server

# Terminal 2: Run client
go run ./client -name "Your Name"
``` 