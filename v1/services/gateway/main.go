// Gateway Service (gateway-svc)
// REST API entrypoint. Connects PG + Redis, runs migrations, enqueues Asynq jobs.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	storage "github.com/lastroseo/services/storage"
	goredis "github.com/redis/go-redis/v9"
)

var (
	pool      *pgxpool.Pool
	asynqCli  *asynq.Client
	redisCli  *goredis.Client
	redisAddr = getEnv("REDIS_ADDR", "localhost:6379")
)

func getEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func main() {
	cfg := storage.ConfigFromEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	pool, err = storage.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("gateway: pg connect: %v", err)
	}
	defer pool.Close()
	log.Println("   PostgreSQL connected")

	if err := storage.Migrate(ctx, pool); err != nil {
		log.Fatalf("gateway: migrate: %v", err)
	}
	log.Println("   Migrations applied")

	asynqCli = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqCli.Close()
	redisCli = goredis.NewClient(&goredis.Options{Addr: redisAddr})
	defer redisCli.Close()
	log.Printf("   Redis (Asynq) connected: %s", redisAddr)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", healthHandler)
	r.Get("/readiness", readinessHandler)
	r.Get("/services", servicesHandler)

	// Also register under /api/ prefix for nginx proxy
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandler)
		r.Get("/readiness", readinessHandler)
		r.Get("/services", servicesHandler)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/projects", listProjectsHandler)
		r.Post("/projects", createProjectHandler)
		r.Get("/projects/{id}", getProjectHandler)
		r.Get("/projects/{id}/keywords", listKeywordsHandler)
		r.Get("/projects/{id}/serp", listSERPHandler)
		r.Get("/projects/{id}/clusters", listClustersHandler)
		r.Get("/projects/{id}/gaps", listGapsHandler)
		r.Post("/projects/{id}/competitors", createCompetitorHandler)
		r.Get("/projects/{id}/competitors", listCompetitorsHandler)
		r.Delete("/competitors/{id}", deleteCompetitorHandler)
		r.Post("/competitors/{id}/inspect", inspectCompetitorHandler)
		r.Get("/jobs/{id}", getJobHandler)
	})

	addr := ":8080"
	log.Printf("🚀 Gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// ── Health ───────────────────────────────────────────────────

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	pgOK := "ok"
	if err := storage.HealthCheck(ctx, pool); err != nil {
		pgOK = "error: " + err.Error()
	}
	redisOK := "ok"
	pong, err := redisCli.Ping(ctx).Result()
	if err != nil {
		redisOK = "error: " + err.Error()
	} else if pong != "PONG" {
		redisOK = "unexpected: " + pong
	}

	status := "ready"
	if pgOK != "ok" || redisOK != "ok" {
		status = "not_ready"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": status,
		"checks": map[string]string{
			"postgres": pgOK,
			"redis":    redisOK,
		},
	})
}

// ── Handlers ─────────────────────────────────────────────────

type createProjectReq struct {
	Name           string   `json:"name"`
	BusinessDesc   string   `json:"business_desc,omitempty"`
	TargetAudience string   `json:"target_audience,omitempty"`
	SeedKeywords   []string `json:"seed_keywords,omitempty"`
}

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	var req createProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json: " + err.Error()})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	ctx := r.Context()
	project, err := storage.CreateProject(ctx, pool, req.Name, req.BusinessDesc, req.TargetAudience, req.SeedKeywords)
	if err != nil {
		log.Printf("ERROR create project: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}

	jobPayload, _ := json.Marshal(map[string]interface{}{
		"project_id":    project.ID,
		"name":          project.Name,
		"business_desc": project.BusinessDesc,
		"seed_keywords": req.SeedKeywords,
	})
	job, err := asynqCli.Enqueue(
		asynq.NewTask("KEYWORD_RESEARCH", jobPayload, asynq.Queue("default")),
	)
	if err != nil {
		log.Printf("ERROR enqueue job: %v", err)
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"project_id": project.ID,
			"project":    project,
			"warning":    "project created but job enqueue failed",
		})
		return
	}

	dbJob, _ := storage.CreateJob(ctx, pool, &storage.Job{
		ProjectID: project.ID,
		Type:      "KEYWORD_RESEARCH",
		Status:    string(storage.JobPending),
		Payload:   jobPayload,
	})

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"project_id": project.ID,
		"project":    project,
		"job_id":     dbJob.ID,
		"asynq_id":   job.ID,
		"status":     "accepted",
	})
}

func listProjectsHandler(w http.ResponseWriter, r *http.Request) {
	projects, err := storage.ListProjects(r.Context(), pool, storage.ListOpts{Limit: 50})
	if err != nil {
		log.Printf("ERROR list projects: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if projects == nil {
		projects = []storage.Project{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"projects": projects,
		"count":    len(projects),
	})
}

func servicesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// SearXNG check
	searxngOK := "ok"
	resp, err := http.Get(getEnv("SEARXNG_URL", "http://searxng:8080") + "/search?q=test&format=json")
	if err != nil || resp.StatusCode != 200 {
		searxngOK = "down"
	}
	if resp != nil {
		resp.Body.Close()
	}

	// PG + Redis
	pgOK := "ok"
	if err := storage.HealthCheck(ctx, pool); err != nil {
		pgOK = "down"
	}
	redisOK := "ok"
	if _, err := redisCli.Ping(ctx).Result(); err != nil {
		redisOK = "down"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"services": map[string]string{
			"gateway":  "ok",
			"postgres": pgOK,
			"redis":    redisOK,
			"searxng":  searxngOK,
		},
	})
}

func getProjectHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	project, err := storage.GetProject(r.Context(), pool, id)
	if err != nil {
		log.Printf("ERROR get project: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if project == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func listKeywordsHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	keywords, err := storage.ListKeywords(r.Context(), pool, id, storage.ListOpts{Limit: 50})
	if err != nil {
		log.Printf("ERROR list keywords: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if keywords == nil {
		keywords = []storage.Keyword{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id": id,
		"keywords":   keywords,
		"count":      len(keywords),
	})
}

func getJobHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := storage.GetJob(r.Context(), pool, id)
	if err != nil {
		log.Printf("ERROR get job: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if job == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func listSERPHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	results, err := storage.ListSERPResultsByProject(r.Context(), pool, id, storage.ListOpts{Limit: 200})
	if err != nil {
		log.Printf("ERROR list serp: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if results == nil {
		results = []storage.SERPProjectResult{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id": id,
		"results":    results,
		"count":      len(results),
	})
}

func listClustersHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	clusters, err := storage.GetClustersByProject(r.Context(), pool, id)
	if err != nil {
		log.Printf("ERROR list clusters: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if clusters == nil {
		clusters = []storage.Cluster{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id": id,
		"clusters":   clusters,
		"count":      len(clusters),
	})
}

func listGapsHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	gaps, err := storage.GetContentGaps(r.Context(), pool, id, storage.ListOpts{Limit: 20})
	if err != nil {
		log.Printf("ERROR list gaps: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if gaps == nil {
		gaps = []storage.ContentGap{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id": id,
		"gaps":       gaps,
		"count":      len(gaps),
	})
}
