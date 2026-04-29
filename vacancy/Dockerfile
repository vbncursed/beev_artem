FROM golang:1.26-alpine AS builder

WORKDIR /src/vacancy

COPY vacancy/go.mod vacancy/go.sum ./
RUN go mod download

COPY vacancy/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/vacancy-service ./cmd/app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/vacancy-service /usr/local/bin/vacancy-service
COPY --from=builder /src/vacancy/config.docker.dev.yaml ./config.docker.dev.yaml
COPY --from=builder /src/vacancy/config.docker.prod.yaml ./config.docker.prod.yaml

EXPOSE 50051

CMD ["vacancy-service"]
