FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY services/google-fetcher/go.mod services/google-fetcher/go.sum ./fetcher/
COPY services/storage ./storage/
WORKDIR /src/fetcher
RUN go mod download 2>/dev/null; true
COPY services/google-fetcher/ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /google-svc .
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /google-svc /google-svc
ENTRYPOINT ["/google-svc"]
