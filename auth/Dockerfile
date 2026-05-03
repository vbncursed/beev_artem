FROM golang:1.26-alpine AS builder

WORKDIR /src/auth

COPY auth/go.mod auth/go.sum ./
RUN go mod download

COPY auth/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/auth-service ./cmd/app && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/admin ./cmd/admin

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/auth-service /usr/local/bin/auth-service
# Operational CLI for role management; used via `docker exec hr-auth admin`.
COPY --from=builder /out/admin /usr/local/bin/admin
COPY --from=builder /src/auth/config.docker.dev.yaml ./config.docker.dev.yaml
COPY --from=builder /src/auth/config.docker.prod.yaml ./config.docker.prod.yaml

EXPOSE 50050

CMD ["auth-service"]
