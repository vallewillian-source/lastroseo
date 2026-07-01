// Package llm provides a lightweight Ollama client for Gemma 4B/9B.
// Used by analytics-svc (intent, cluster naming, content gaps)
// and extractor-svc (page topic extraction).
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client wraps Ollama's /api/generate endpoint.
type Client struct {
	baseURL string
	model   string
	http    *http.Client
}

// NewClient creates an Ollama client. Default model is gemma4:e4b.
func NewClient(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "gemma4:e4b"
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

type generateReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResp struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Generate sends a prompt and returns the trimmed response.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(generateReq{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("llm: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm: generate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm: status %d: %s", resp.StatusCode, string(b))
	}

	var r generateResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("llm: decode: %w", err)
	}
	return strings.TrimSpace(r.Response), nil
}

// ── Prompt templates ──────────────────────────────────────────

// ClassifyIntent prompts Gemma to classify search intent.
func ClassifyIntent(ctx context.Context, c *Client, keyword, businessDesc string) (string, error) {
	prompt := fmt.Sprintf(`Classifique a intenção de busca desta keyword em UMA palavra:
- informacional (quer aprender: como, o que é, guia, tutorial, dicas)
- transacional (quer comprar/contratar: preço, plano, orçamento, comprar)
- comercial (quer comparar: melhor, vs, review, vale a pena, alternativas)
- navegacional (quer acessar: login, site, app, download, telefone, endereço)

Contexto do negócio: %s

Keyword: "%s"

Responda apenas uma palavra.`, businessDesc, keyword)

	resp, err := c.Generate(ctx, prompt)
	if err != nil {
		return "informacional", err
	}
	resp = strings.ToLower(strings.TrimSpace(resp))
	for _, intent := range []string{"informacional", "transacional", "comercial", "navegacional"} {
		if strings.Contains(resp, intent) {
			return intent, nil
		}
	}
	return "informacional", nil
}

// NameCluster prompts Gemma to name a keyword cluster.
func NameCluster(ctx context.Context, c *Client, keywords []string) (string, error) {
	joined := strings.Join(keywords, ", ")
	if len(joined) > 800 {
		joined = joined[:800]
	}
	prompt := fmt.Sprintf(`Dê um nome curto (2-4 palavras, em português) para este grupo de keywords relacionadas.
O nome deve capturar o tema comum. Seja específico, não genérico.

Keywords: %s

Responda apenas o nome.`, joined)

	return c.Generate(ctx, prompt)
}

// DetectContentGaps prompts Gemma to find content gaps vs competitors.
func DetectContentGaps(ctx context.Context, c *Client, keyword string, competitorTopics []string) (string, error) {
	joined := strings.Join(competitorTopics, "\n- ")
	if len(joined) > 1200 {
		joined = joined[:1200]
	}
	prompt := fmt.Sprintf(`Analise os tópicos que concorrentes cobrem para a keyword "%s".
Liste 3-5 tópicos ou ângulos que NÃO estão sendo cobertos (gaps de conteúdo).
Seja específico — sugira títulos de seções ou subtópicos concretos.

Tópicos dos concorrentes:
- %s

Responda em português, um gap por linha, formato: "- [gap]"`, keyword, joined)

	return c.Generate(ctx, prompt)
}

// ExtractTopics prompts Gemma to extract main topics from page content.
func ExtractTopics(ctx context.Context, c *Client, title, h1, h2s, snippet string) (string, error) {
	prompt := fmt.Sprintf(`Extraia os 3-5 principais tópicos/entidades desta página web.
Responda em português, um tópico por linha, no formato: "- tópico".

Título: %s
H1: %s
H2s: %s
Snippet: %s`, title, h1, truncate(h2s, 500), truncate(snippet, 300))

	return c.Generate(ctx, prompt)
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
