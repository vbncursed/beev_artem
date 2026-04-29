FROM golang:1.26-alpine AS builder

WORKDIR /src/analysis

COPY analysis/go.mod analysis/go.sum ./
RUN go mod download

COPY analysis/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/analysis-service ./cmd/app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/analysis-service /usr/local/bin/analysis-service
COPY --from=builder /src/analysis/config.docker.dev.yaml ./config.docker.dev.yaml
COPY --from=builder /src/analysis/config.docker.prod.yaml ./config.docker.prod.yaml

EXPOSE 50054

CMD ["analysis-service"]
