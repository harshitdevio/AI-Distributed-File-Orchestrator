    FROM golang:1.25-alpine AS builder

    RUN apk add --no-cache git

    WORKDIR /app

    COPY go.mod go.sum ./
    RUN go mod download

    COPY . .

    RUN CGO_ENABLED=0 GOOS=linux go build -o /go-service .

    FROM alpine:latest

    RUN adduser -D appuser
    USER appuser

    WORKDIR /

    COPY --from=builder /go-service /go-service

    ENTRYPOINT ["/go-service"]

    CMD ["/go-service"]