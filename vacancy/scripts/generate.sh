#!/bin/bash

cd "$(dirname "$0")/.." || exit

mkdir -p ./internal/pb ./internal/pb/swagger

GATEWAY_PATH=$(go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2)

protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --go_out=./internal/pb --go_opt=paths=source_relative \
  --go-grpc_out=./internal/pb --go-grpc_opt=paths=source_relative \
  ./api/common/common.proto ./api/models/vacancy_model.proto ./api/vacancy_api/vacancy.proto

protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --grpc-gateway_out=./internal/pb \
  --grpc-gateway_opt paths=source_relative \
  --grpc-gateway_opt logtostderr=true \
  ./api/vacancy_api/vacancy.proto

protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --openapiv2_out=./internal/pb/swagger \
  --openapiv2_opt logtostderr=true \
  ./api/vacancy_api/vacancy.proto
