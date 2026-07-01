FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY services/analytics/go.mod services/analytics/go.sum* ./analytics/
COPY services/storage ./storage/
COPY pkg/llm ./pkg/llm/
WORKDIR /src/analytics
RUN go mod download 2>/dev/null; true
COPY services/analytics/ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /analytics-svc .
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /analytics-svc /analytics-svc
ENTRYPOINT ["/analytics-svc"]
