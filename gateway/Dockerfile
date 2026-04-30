FROM golang:1.26-alpine AS builder

WORKDIR /src/gateway

COPY gateway/go.mod gateway/go.sum ./
RUN go mod download

COPY gateway/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gateway-service ./cmd/app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/gateway-service /usr/local/bin/gateway-service
COPY --from=builder /src/gateway/config.docker.dev.yaml ./config.docker.dev.yaml
COPY --from=builder /src/gateway/config.docker.prod.yaml ./config.docker.prod.yaml
COPY --from=builder /src/gateway/internal/pb/openapi ./internal/pb/openapi

EXPOSE 8080

CMD ["gateway-service"]
