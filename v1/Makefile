.PHONY: build up down infra app logs clean

# Build all 5 Go services
build:
	cd services/gateway && CGO_ENABLED=0 go build -o gateway-svc .
	cd services/google-fetcher && CGO_ENABLED=0 go build -o google-svc .
	cd services/crawler && CGO_ENABLED=0 go build -o crawler-svc .
	cd services/extractor && CGO_ENABLED=0 go build -o extractor-svc .
	cd services/analytics && CGO_ENABLED=0 go build -o analytics-svc .

# Infrastructure only (postgres, redis, searxng, ollama)
infra:
	docker-compose up -d postgres redis searxng ollama

# App services only (after build)
app:
	docker-compose --profile app up -d

# Everything: infra + app
up: infra build app

# Stop all
down:
	docker-compose --profile app down
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Clean built binaries
clean:
	rm -f services/gateway/gateway-svc
	rm -f services/google-fetcher/google-svc
	rm -f services/crawler/crawler-svc
	rm -f services/extractor/extractor-svc
	rm -f services/analytics/analytics-svc
