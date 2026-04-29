FROM golang:1.26-alpine AS builder

WORKDIR /src/resume

COPY resume/go.mod resume/go.sum ./
RUN go mod download

COPY resume/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/resume-service ./cmd/app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/resume-service /usr/local/bin/resume-service
COPY --from=builder /src/resume/config.docker.dev.yaml ./config.docker.dev.yaml
COPY --from=builder /src/resume/config.docker.prod.yaml ./config.docker.prod.yaml

EXPOSE 50052

CMD ["resume-service"]
