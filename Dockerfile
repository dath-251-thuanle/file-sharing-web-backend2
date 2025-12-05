FROM golang:1.25.4-alpine AS builder

RUN apk update && apk add --no-cache git curl

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/server

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/main .

COPY .env .

EXPOSE 8080

CMD ["./main"]