#!/bin/bash

# Exit on any error
set -e

echo "Setting up Starbit server environment..."

# Update and install packages
sudo apt-get update && sudo apt-get upgrade -y
sudo apt-get install -y protobuf-compiler git wget snapd

# Install Go and set up environment
sudo snap install go --classic --channel=1.23/stable

# Set up Go environment in .profile (this gets sourced by .bashrc)
echo 'export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:/snap/bin:$GOBIN' >> ~/.profile

# Create Go workspace and set up current session
mkdir -p $HOME/go/bin
source ~/.profile

# Verify installation
go version

# Install and verify protobuf plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
which protoc-gen-go

# Generate protobuf files and make server executable
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/service.proto
chmod +x run_server.sh

# Configure UDP buffer size for QUIC
echo "net.core.rmem_max=2500000
net.core.wmem_max=2500000" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

echo "Setup complete! You can now run ./run_server.sh"
echo "To keep the server running after disconnecting from SSH, use:"
echo "nohup ./run_server.sh > server.log 2>&1 &"
