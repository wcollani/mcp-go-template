FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add -U --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /mcp-server ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /mcp-server /mcp-server

EXPOSE 8080

ENTRYPOINT ["/mcp-server"]
