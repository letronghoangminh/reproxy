FROM golang:1.24-alpine AS builder

    WORKDIR /app

    COPY go.mod go.sum ./
    RUN go mod download

    COPY . .

    RUN CGO_ENABLED=0 GOOS=linux go build -o reproxy ./cmd/main.go

FROM alpine:latest

    WORKDIR /app

    COPY --from=builder /app/reproxy .
    COPY --from=builder /app/config ./config/

    EXPOSE 2209

    CMD ["./reproxy", "--config", "config/config.yaml"]
