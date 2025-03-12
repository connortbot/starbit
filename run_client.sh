#! /bin/bash

LOG_FILE="debug${1}.log"
echo "Clearing ${LOG_FILE}..."
> "${LOG_FILE}"

echo "Running client with USER=$1"
USER="$1" DEBUG=1 go run ./client