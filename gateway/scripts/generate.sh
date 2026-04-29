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

mkdir -p ./internal/pb/swagger
protoc -I ./api \
  -I ./api/google/api \
  -I "${GATEWAY_PATH}" \
  --openapiv2_out=./internal/pb/swagger \
  --openapiv2_opt logtostderr=true \
  ./api/auth_api/auth.proto \
  ./api/vacancy_api/vacancy.proto \
  ./api/resume_api/resume.proto \
  ./api/analysis_api/analysis.proto
