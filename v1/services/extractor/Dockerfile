FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY services/extractor/go.mod services/extractor/go.sum ./extractor/
COPY services/storage ./storage/
COPY pkg/llm ./pkg/llm/
WORKDIR /src/extractor
RUN go mod download 2>/dev/null; true
COPY services/extractor/ .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /extractor-svc .
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /extractor-svc /extractor-svc
ENTRYPOINT ["/extractor-svc"]
