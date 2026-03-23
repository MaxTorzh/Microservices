FROM golang:1.26-alpine AS builder

RUN apk add --no-cache protoc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    export PATH="$PATH:$(go env GOPATH)/bin" && \
    protoc --proto_path=api/proto \
           --go_out=. \
           --go_opt=paths=source_relative \
           --go-grpc_out=. \
           --go-grpc_opt=paths=source_relative \
           api/proto/user/user.proto

RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

COPY --from=builder /app/user-service .

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./user-service"]