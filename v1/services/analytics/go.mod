module github.com/lastroseo/services/analytics

go 1.23

require (
	github.com/hibiken/asynq v0.26.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/lastroseo/pkg/llm v0.0.0
	github.com/lastroseo/services/storage v0.0.0
	github.com/redis/go-redis/v9 v9.14.1
)

replace github.com/lastroseo/services/storage => ../storage

replace github.com/lastroseo/pkg/llm => ../pkg/llm
