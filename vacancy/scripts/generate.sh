#!/bin/bash

cd "$(dirname "$0")/.." || exit

mkdir -p ./internal/pb

GATEWAY_PATH=$(go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2)

protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --go_out=./internal/pb --go_opt=paths=source_relative \
  --go-grpc_out=./internal/pb --go-grpc_opt=paths=source_relative \
  ./api/common/common.proto ./api/models/vacancy_model.proto ./api/vacancy_api/vacancy.proto ./api/auth_api/auth.proto

protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --grpc-gateway_out=./internal/pb \
  --grpc-gateway_opt paths=source_relative \
  --grpc-gateway_opt logtostderr=true \
  ./api/vacancy_api/vacancy.proto

# OpenAPI emission lives in the gateway service: it owns the public docs
# surface, generates a single merged OpenAPI 3.0 document via gnostic, and
# serves it at /swagger.json + /docs. Backend services do not emit swagger
# files because nobody consumes them.
