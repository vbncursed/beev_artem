#!/bin/bash

cd "$(dirname "$0")/.." || exit

mkdir -p ./internal/pb

protoc -I ./api \
  --go_out=./internal/pb --go_opt=paths=source_relative \
  --go-grpc_out=./internal/pb --go-grpc_opt=paths=source_relative \
  ./api/multiagent_api/multiagent.proto
