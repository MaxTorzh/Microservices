FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk update && apk upgrade && \
    apk add --no-cache --update ca-certificates

RUN apk add --no-cache git || \
    (sleep 5 && apk add --no-cache git) || \
    (sleep 10 && apk add --no-cache git)

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

COPY --from=builder /app/notification-service .

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

CMD ["./notification-service"]