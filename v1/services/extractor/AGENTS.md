# Content Extractor — AGENTS.md

> **Asynq worker. HTML download + regex parsing + LLM topic extraction.**

## Role
Consumes `CONTENT_EXTRACTION` jobs. Downloads HTML from URLs, parses metadata via regex (no external HTML parser), extracts topics via Gemma LLM, and stores in `page_data` + `keyword_pages`.

## Pipeline (per URL)

```
1. HTTP GET → max 1MB, 15s timeout
2. Regex extraction:
   - <title> → title
   - <meta name="description" content="..."> → meta_description
   - <h1> → h1 array
   - <h2> → h2 array
   - <h3> → h3 array
   - word count (text content)
   - image count (<img>)
   - video count (<video> + youtube embeds)
3. LLM topic extraction (goroutine, best-effort):
   → Gemma prompt: "Extraia 3-5 tópicos principais desta página"
   → Stores in page_data.topics
4. Upsert page_data
5. Insert keyword_page relationship (keyword → page, position)
```

## Key implementation

```go
// Regex-based parsing — zero external dependencies
func extractTitle(html string) string
func extractMetaDescription(html string) string
func extractHeadings(html, tag string) []string
func countWords(html string) int
func countMatches(html, pattern string) int

// LLM (non-blocking)
go func() {
    topics := llm.ExtractTopics(ctx, client, title, h1, h2s, snippet)
    pool.Exec(ctx, "UPDATE page_data SET topics=$1 WHERE id=$2", topics, pageID)
}()
```

## Output tables
- `page_data` — title, meta_description, h1/h2/h3, word_count, image_count, video_count, topics
- `keyword_pages` — keyword_id, page_id, position

## What is NOT implemented
- goquery HTML parser → regex is simpler, no CGO needed
- TF-IDF extraction → LLM topics provide better semantics
- Double Inverted Index → keyword_pages table handles JOINs
- Ollama local instance → uses host desktop Ollama (172.17.0.1:11434)

## Dependencies
- `github.com/hibiken/asynq`
- `github.com/jackc/pgx/v5`
- `github.com/lastroseo/pkg/llm` — Gemma 4B client
- `github.com/lastroseo/services/storage`
