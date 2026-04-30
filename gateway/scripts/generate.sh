#!/bin/bash

cd "$(dirname "$0")/.." || exit

mkdir -p ./internal/pb ./internal/pb/openapi

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

# Per-service OpenAPI 3.0 specs via gnostic. One invocation per service so
# Scalar UI can render a dropdown with each service as a separate source.
# gnostic always names its output "openapi.yaml" — we emit each into a temp
# subdir and rename to "<svc>.yaml" so the runtime resolves auth.yaml /
# vacancy.yaml / etc. on its own URL.
gen_openapi() {
  local svc="$1"
  local title="$2"
  local proto="$3"
  local tmp="./internal/pb/openapi/.${svc}_tmp"
  rm -rf "${tmp}"
  mkdir -p "${tmp}"
  protoc -I ./api \
    -I ./api/google/api \
    -I "${GATEWAY_PATH}" \
    --openapi_out=title="${title}",version="1.0.0":"${tmp}" \
    "${proto}"
  mv "${tmp}/openapi.yaml" "./internal/pb/openapi/${svc}.yaml"
  rmdir "${tmp}"
}

# Drop stale specs first so a renamed/removed service doesn't leak its old
# yaml into the dropdown.
rm -f ./internal/pb/openapi/*.yaml

gen_openapi auth     "Auth API"     ./api/auth_api/auth.proto
gen_openapi vacancy  "Vacancy API"  ./api/vacancy_api/vacancy.proto
gen_openapi resume   "Resume API"   ./api/resume_api/resume.proto
gen_openapi analysis "Analysis API" ./api/analysis_api/analysis.proto
