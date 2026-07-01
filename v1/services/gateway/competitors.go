package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lastroseo/services/storage"
)

// ── Competitor CRUD Handlers ────────────────────────────────

type createCompetitorReq struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func createCompetitorHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	var req createCompetitorReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json: " + err.Error()})
		return
	}
	if req.Name == "" || req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and url are required"})
		return
	}

	c, err := storage.CreateCompetitor(r.Context(), pool, projectID, req.Name, req.URL)
	if err != nil {
		log.Printf("ERROR create competitor: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func listCompetitorsHandler(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	competitors, err := storage.ListCompetitors(r.Context(), pool, projectID)
	if err != nil {
		log.Printf("ERROR list competitors: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if competitors == nil {
		competitors = []storage.Competitor{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"competitors": competitors,
		"count":       len(competitors),
	})
}

func deleteCompetitorHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := storage.DeleteCompetitor(r.Context(), pool, id); err != nil {
		log.Printf("ERROR delete competitor: %v", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ── Inspect Endpoint ────────────────────────────────────────

type inspectResult struct {
	Keyword string  `json:"keyword"`
	Score   float64 `json:"score"`
	Freq    int     `json:"count"`
}

func inspectCompetitorHandler(w http.ResponseWriter, r *http.Request) {
	competitorID := chi.URLParam(r, "id")

	comp, err := storage.GetCompetitor(r.Context(), pool, competitorID)
	if err != nil {
		log.Printf("ERROR get competitor: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "database error"})
		return
	}
	if comp == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "competitor not found"})
		return
	}

	// 1. Fetch HTML
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", comp.URL, nil)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid URL: " + err.Error()})
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; LastroSEO/1.0; +https://lastroseo.com)")
	req.Header.Set("Accept", "text/html")

	httpResp, err := (&http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}).Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "fetch failed: " + err.Error()})
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": fmt.Sprintf("HTTP %d from %s", httpResp.StatusCode, comp.URL)})
		return
	}

	// 2. Read HTML
	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 2<<20)) // 2 MB max
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "reading response: " + err.Error()})
		return
	}

	// 3. Extract SEO-focused text (title, meta, headings, first paragraphs)
	seoText := extractSEOText(string(body))
	if strings.TrimSpace(seoText) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "no SEO-relevant content found on page"})
		return
	}

	// 4. Get project business_desc for Gemma4 context
	project, _ := storage.GetProject(ctx, pool, comp.ProjectID)
	bizDesc := ""
	if project != nil {
		bizDesc = project.BusinessDesc
	}

	// 5. Call Gemma4 to extract keywords directly from the SEO text
	keywords, err := askGemmaClassify(ctx, seoText, bizDesc)
	if err != nil {
		log.Printf("ERROR gemma classify: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Gemma4 analysis failed: " + err.Error()})
		return
	}

	// If no keywords were classified, ensure we return empty slice (not null JSON)
	if keywords == nil {
		keywords = []SEOKeyword{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"competitor_id": competitorID,
		"name":          comp.Name,
		"url":           comp.URL,
		"keywords":      keywords,
		"count":         len(keywords),
	})
}
