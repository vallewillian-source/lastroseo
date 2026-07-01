FROM golang:1.25-bookworm AS builder
WORKDIR /src

# Copy gateway module
COPY services/gateway/go.mod services/gateway/go.sum ./gateway/
COPY services/storage ./storage/

# Get gateway deps
WORKDIR /src/gateway
RUN go mod download 2>/dev/null; true

# Copy gateway source
COPY services/gateway/ .

# Build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway-svc .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /gateway-svc /gateway-svc
EXPOSE 8080
ENTRYPOINT ["/gateway-svc"]
