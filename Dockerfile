FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/server



FROM alpine:3.19

RUN apk add --no-cache \
    ca-certificates \
    curl \
    postgresql-client

WORKDIR /app

RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then BIN=amd64; \
    elif [ "$ARCH" = "aarch64" ]; then BIN=arm64; \
    else echo "unsupported arch $ARCH" && exit 1; fi && \
    curl -fL https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-$BIN.tar.gz \
    | tar -xz && \
    mv migrate /usr/local/bin/migrate && chmod +x /usr/local/bin/migrate

COPY --from=builder /app/main ./main
COPY internal/infrastructure/database /app/migrations
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
