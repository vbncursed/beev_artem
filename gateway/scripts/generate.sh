#!/bin/bash

cd "$(dirname "$0")/.." || exit

GATEWAY_PATH=$(go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2)

protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --go_out=./internal/pb --go_opt=paths=source_relative \
  --go-grpc_out=./internal/pb --go-grpc_opt=paths=source_relative \
  ./api/common/common.proto \
  ./api/models/auth_model.proto \
  ./api/models/vacancy_model.proto \
  ./api/models/resume_model.proto \
  ./api/models/analysis_model.proto \
  ./api/auth_api/auth.proto \
  ./api/vacancy_api/vacancy.proto \
  ./api/resume_api/resume.proto \
  ./api/analysis_api/analysis.proto

protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --grpc-gateway_out=./internal/pb \
  --grpc-gateway_opt paths=source_relative \
  --grpc-gateway_opt logtostderr=true \
  ./api/auth_api/auth.proto \
  ./api/vacancy_api/vacancy.proto \
  ./api/resume_api/resume.proto \
  ./api/analysis_api/analysis.proto

# OpenAPI 3.0 via gnostic's protoc-gen-openapi. One invocation across all
# four service protos produces a single merged ./internal/pb/openapi/openapi.yaml,
# so the gateway just reads + JSON-marshals it at boot — no per-service
# merge logic.
mkdir -p ./internal/pb/openapi
protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --openapi_out=title="Gateway API",version="1.0.0",description="Unified gateway API docs":./internal/pb/openapi \
  ./api/auth_api/auth.proto \
  ./api/vacancy_api/vacancy.proto \
  ./api/resume_api/resume.proto \
  ./api/analysis_api/analysis.proto
