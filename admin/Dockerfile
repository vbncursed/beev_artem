FROM golang:1.26-alpine AS builder

WORKDIR /src/admin

COPY admin/go.mod admin/go.sum ./
RUN go mod download

COPY admin/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/admin-service ./cmd/app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/admin-service /usr/local/bin/admin-service
COPY --from=builder /src/admin/config.docker.dev.yaml ./config.docker.dev.yaml
COPY --from=builder /src/admin/config.docker.prod.yaml ./config.docker.prod.yaml

EXPOSE 50056

CMD ["admin-service"]
