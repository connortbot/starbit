#!/bin/bash

# Exit on any error
set -e

echo "Setting up Starbit server environment..."

# Update system packages
echo "Updating system packages..."
sudo apt-get update
sudo apt-get upgrade -y

# Install required packages
echo "Installing required packages..."
sudo apt-get install -y protobuf-compiler git wget

# Install Go 1.23
echo "Installing Go 1.23..."
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
rm go1.23.0.linux-amd64.tar.gz

# Set up Go environment
echo "Setting up Go environment..."
echo 'export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin

# Verify Go installation
go version

# Install Go protobuf plugins
echo "Installing protobuf plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf files
echo "Generating protobuf files..."
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/service.proto

# Make run_server.sh executable
echo "Making run_server.sh executable..."
chmod +x run_server.sh

# Increase UDP buffer size for better QUIC performance
echo "Increasing UDP buffer size..."
sudo sysctl -w net.core.rmem_max=2500000
sudo sysctl -w net.core.wmem_max=2500000

echo "Setup complete! You can now run ./run_server.sh"
echo "To keep the server running after disconnecting from SSH, use:"
echo "nohup ./run_server.sh > server.log 2>&1 &"
