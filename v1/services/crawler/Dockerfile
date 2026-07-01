FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY services/crawler/go.mod services/crawler/go.sum ./crawler/
COPY services/storage ./storage/
WORKDIR /src/crawler
RUN go mod download 2>/dev/null; true
COPY services/crawler/ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /crawler-svc .
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /crawler-svc /crawler-svc
ENTRYPOINT ["/crawler-svc"]
