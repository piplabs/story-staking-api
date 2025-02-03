FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o staking-api ./cmd/main.go

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

RUN adduser -D -u 1000 appuser
USER appuser

COPY --from=builder --chown=appuser:appuser /app/staking-api /app/staking-api

WORKDIR /app

EXPOSE 8080

ENTRYPOINT ["/app/staking-api"]
