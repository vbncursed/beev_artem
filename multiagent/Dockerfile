FROM golang:1.26-alpine AS builder

WORKDIR /src/multiagent

COPY multiagent/go.mod multiagent/go.sum ./
RUN go mod download

COPY multiagent/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/multiagent-service ./cmd/app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/multiagent-service /usr/local/bin/multiagent-service
COPY --from=builder /src/multiagent/config.docker.dev.yaml ./config.docker.dev.yaml
COPY --from=builder /src/multiagent/config.docker.prod.yaml ./config.docker.prod.yaml

EXPOSE 50055

CMD ["multiagent-service"]
