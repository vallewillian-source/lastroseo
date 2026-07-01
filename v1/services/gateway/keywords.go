package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
)

// ── Types ───────────────────────────────────────────────────

type SEOKeyword struct {
	Keyword string  `json:"keyword"`
	Score   float64 `json:"score"`
	Freq    int     `json:"count"`
}

type candidate struct {
	ngram string
	freq  int
}

// ── Step 1: Focused HTML extraction ─────────────────────────

// extractSEOText parses HTML and extracts only SEO-relevant text:
// title, meta description, h1-h6, and first few paragraphs.
// Discards nav, footer, sidebar, scripts, styles, etc.
// Title and headings are repeated for weighting.
func extractSEOText(raw string) string {
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return raw
	}

	// Tags to strip entirely (including content)
	stripTags := map[string]bool{
		"script": true, "style": true, "noscript": true, "iframe": true,
		"svg": true, "path": true, "head": true, "link": true, "meta": true,
		"form": true, "input": true, "button": true, "select": true, "textarea": true,
		"canvas": true, "video": true, "audio": true, "embed": true, "object": true,
		"applet": true, "frameset": true, "frame": true, "noembed": true,
	}

	// Tags that are noise — skip their content entirely
	noiseTags := map[string]bool{
		"nav": true, "footer": true, "aside": true, "header": true,
		"ul": true, "ol": true, // lists are usually navigation
	}

	// Track what we've extracted
	var title, metaDesc string
	var headings, paragraphs []string
	paraCount := 0
	maxParas := 5 // increased from 3

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			tag := strings.ToLower(n.Data)

			// Extract <title>
			if tag == "title" {
				title = extractTextContent(n)
				return
			}

			// Extract <meta name="description">
			if tag == "meta" {
				name := getAttr(n, "name")
				if strings.ToLower(name) == "description" {
					metaDesc = getAttr(n, "content")
				}
				return
			}

			if stripTags[tag] {
				return
			}
			if noiseTags[tag] {
				return
			}

			// Extract headings
			if tag == "h1" || tag == "h2" || tag == "h3" || tag == "h4" || tag == "h5" || tag == "h6" {
				t := extractTextContent(n)
				if t != "" {
					headings = append(headings, t)
				}
				return
			}

			// Extract paragraphs (limited)
			if tag == "p" && paraCount < maxParas {
				t := extractTextContent(n)
				if t != "" && len(t) > 20 { // skip tiny paragraphs
					paragraphs = append(paragraphs, t)
					paraCount++
				}
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	// Build focused text with weighting via repetition:
	// Title x5, headings x3, meta x2, paragraphs x1
	var b strings.Builder
	if title != "" {
		for i := 0; i < 5; i++ {
			b.WriteString(title)
			b.WriteString(" ")
		}
	}
	if metaDesc != "" {
		for i := 0; i < 2; i++ {
			b.WriteString(metaDesc)
			b.WriteString(" ")
		}
	}
	for _, h := range headings {
		for i := 0; i < 3; i++ {
			b.WriteString(h)
			b.WriteString(" ")
		}
	}
	for _, p := range paragraphs {
		b.WriteString(p)
		b.WriteString(" ")
	}

	return b.String()
}

// extractTextContent gets all text from a node and its children.
func extractTextContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			t := strings.TrimSpace(node.Data)
			if t != "" {
				b.WriteString(t)
				b.WriteString(" ")
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}

// getAttr returns an attribute value from an HTML node.
func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

// ── Step 2: N-gram generation ───────────────────────────────

// generateNGrams tokenizes text and produces 2-4 word n-grams,
// filtering out those that are all stop words.
func generateNGrams(text string, minN, maxN int) []candidate {
	text = strings.ToLower(text)

	// Tokenize: split on non-letter/non-number, keep accented chars
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	// Filter out stop words and very short words
	var filtered []string
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) < 2 || isStopWord(w) {
			continue
		}
		filtered = append(filtered, w)
	}

	if len(filtered) < minN {
		return nil
	}

	// Count n-gram frequencies
	freq := map[string]int{}
	for n := minN; n <= maxN; n++ {
		for i := 0; i <= len(filtered)-n; i++ {
			ngram := strings.Join(filtered[i:i+n], " ")
			freq[ngram]++
		}
	}

	// Collect candidates with adaptive threshold:
	// - bigrams/trigrams: freq >= 1 (after weighting, they appear at least once)
	// - 4-grams: freq >= 2 (only if they appear multiple times)
	// Filter out n-grams that are pure navigation noise
	var candidates []candidate
	seen := map[string]bool{}
	for ngram, f := range freq {
		if f < 1 {
			continue
		}
		if seen[ngram] {
			continue
		}
		words := strings.Fields(ngram)
		n := len(words)

		// For 4-grams, require freq >= 2
		if n >= 4 && f < 2 {
			continue
		}

		// Skip pure navigation/boilerplate n-grams
		navigationWords := map[string]bool{
			"cookies": true, "privacidade": true, "política": true, "termos": true,
			"solucões": true, "segmentos": true, "ferramentas": true,
		}
		allNav := true
		for _, w := range words {
			if !navigationWords[w] {
				allNav = false
				break
			}
		}
		if allNav && n >= 3 {
			continue
		}

		seen[ngram] = true
		candidates = append(candidates, candidate{ngram: ngram, freq: f})
	}

	// Sort by frequency descending, then by shorter n-grams first
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].freq != candidates[j].freq {
			return candidates[i].freq > candidates[j].freq
		}
		return len(candidates[i].ngram) < len(candidates[j].ngram)
	})

	return candidates
}

// ── Step 3: Ollama / Gemma4 call ────────────────────────────

type ollamaGenReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaGenResp struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// askGemmaClassify sends the SEO text to Gemma4 and asks it to directly
// list SEO keywords. Simpler than classification — lets the model generate.
func askGemmaClassify(ctx context.Context, seoText string, businessDesc string) ([]SEOKeyword, error) {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://172.17.0.1:11434"
	}

	// Truncate text to fit in context
	text := seoText
	if len(text) > 1500 {
		text = text[:1500]
	}

	biz := businessDesc
	if len(biz) > 200 {
		biz = biz[:200]
	}

	prompt := fmt.Sprintf(`Voce e um especialista em SEO. Abaixo esta o texto principal de uma pagina web sobre: %s

Extraia de 10 a 15 palavras-chave que um usuario pesquisaria no Google para encontrar esta pagina.

Texto da pagina:
%s

Responda APENAS uma palavra-chave por linha, sem numeracao, sem explicacao, sem aspas.
Exemplo de saida:
chatbot whatsapp
atendimento automatizado
crm para vendas`, biz, text)

	// Call Ollama
	body, _ := json.Marshal(ollamaGenReq{
		Model:  "gemma4:e4b",
		Prompt: prompt,
		Stream: false,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", ollamaURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gemma: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemma: connect failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemma: status %d: %s", resp.StatusCode, string(b))
	}

	var ollamaResp ollamaGenResp
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("gemma: decode: %w", err)
	}

	// Parse response: expect one keyword per line
	lines := strings.Split(strings.TrimSpace(ollamaResp.Response), "\n")
	var results []SEOKeyword
	seen := map[string]bool{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "0123456789. ")
		line = strings.TrimSpace(line)
		line = strings.Trim(line, "\"'")

		if line == "" || len(line) < 3 {
			continue
		}

		// Skip lines that are instructions or examples
		lower := strings.ToLower(line)
		if strings.Contains(lower, "exemplo") || strings.Contains(lower, "resposta") ||
			strings.Contains(lower, "saida") || strings.Contains(lower, "nota") {
			continue
		}

		// Deduplicate
		if seen[lower] {
			continue
		}
		seen[lower] = true

		results = append(results, SEOKeyword{
			Keyword: line,
			Score:   1.0,
			Freq:    1,
		})
	}

	return results, nil
}

// ── Stop words (kept from previous implementation) ──────────

var stopWords map[string]bool

func init() {
	seen := map[string]bool{}
	stopWords = make(map[string]bool)
	for _, w := range stopWordsList {
		if !seen[w] {
			stopWords[w] = true
			seen[w] = true
		}
	}
}

var stopWordsList = []string{
	// English
	"a", "about", "above", "after", "again", "against", "all", "am", "an",
	"and", "any", "are", "aren't", "as", "at", "be", "because", "been",
	"before", "being", "below", "between", "both", "but", "by", "can't",
	"cannot", "could", "couldn't", "did", "didn't", "do", "does", "doesn't",
	"doing", "don't", "down", "during", "each", "few", "for", "from",
	"further", "got", "had", "hadn't", "has", "hasn't", "have", "haven't",
	"having", "he", "her", "here", "hers", "herself", "him", "himself",
	"his", "how", "i", "i'd", "i'll", "i'm", "i've", "if", "in", "into",
	"is", "isn't", "it", "its", "itself", "just", "let", "me", "more",
	"most", "mustn't", "my", "myself", "no", "nor", "not", "now", "of",
	"off", "on", "once", "only", "or", "other", "ought", "our", "ours",
	"ourselves", "out", "over", "own", "same", "shan't", "she", "she'd",
	"she'll", "she's", "should", "shouldn't", "so", "some", "such", "than",
	"that", "that's", "the", "their", "theirs", "them", "themselves", "then",
	"there", "there's", "these", "they", "they'd", "they'll", "they're",
	"they've", "this", "those", "through", "to", "too", "under", "until",
	"up", "very", "was", "wasn't", "we", "we'd", "we'll", "we're", "we've",
	"were", "weren't", "what", "what's", "when", "when's", "where", "where's",
	"which", "while", "who", "who's", "whom", "why", "why's", "will", "with",
	"won't", "would", "wouldn't", "you", "you'd", "you'll", "you're", "you've",
	"your", "yours", "yourself", "yourselves",
	// Portuguese
	"à", "ao", "aos", "às", "com", "como", "da", "das", "de", "del", "dem",
	"depois", "do", "dos", "é", "ela", "elas", "ele", "eles", "em", "entre",
	"era", "essa", "essas", "esse", "esses", "esta", "estas", "este", "estes",
	"eu", "fui", "há", "ir", "isso", "isto", "já", "lhe", "lhes", "lo",
	"mais", "mas", "me", "mesmo", "meu", "meus", "minha", "minhas", "muito",
	"muitos", "na", "nas", "não", "nem", "no", "nos", "nós", "num", "numa",
	"o", "os", "ou", "para", "pela", "pelas", "pelo", "pelos", "por", "qual",
	"quando", "que", "quem", "são", "se", "sem", "ser", "seu", "seus", "sim",
	"só", "também", "te", "tem", "teu", "teus", "tu", "tua", "tuas", "um",
	"uma", "uns", "umas", "você", "vocês", "vos", "foram", "será", "tenho",
	"tinha", "ter", "tenha", "fosse", "tivesse", "sido", "está", "estava",
	"estou", "estiver", "estivesse", "estivera", "esteja", "teria",
	// Common HTML / web noise words
	"click", "read", "new", "post", "blog", "news", "home", "page", "site",
	"web", "www", "http", "https", "com", "org", "net", "br", "div", "span",
	"class", "style", "href", "src", "alt", "title", "type", "name", "value",
	"id", "role", "aria", "data", "icon", "image", "img", "video", "menu",
	"nav", "footer", "header", "main", "content", "text", "button", "form",
	"input", "cookie", "privacy", "policy", "terms", "use", "using", "used",
	"can", "may", "might", "way", "well", "also", "like", "make", "many",
	"much", "great", "good", "best", "top", "one", "two", "three", "first",
	"last", "next", "back", "see", "look", "find", "help", "support",
	"contact", "about", "us", "our", "team", "job", "jobs", "career", "apply",
	"learn", "get", "take", "try", "start", "free", "subscribe", "newsletter",
	"sign", "share", "follow", "social", "media", "twitter", "facebook",
	"instagram", "linkedin", "youtube", "close", "open", "show", "hide",
	"search", "results", "loading", "error", "success", "true", "false",
	"null", "undefined", "function", "return", "var", "let", "const",
	"saiba", "mais", "clique", "aqui", "todos", "direitos", "reservados",
	"copyright", "política", "privacidade", "termos", "uso", "cookies",
}

func isStopWord(word string) bool {
	return stopWords[word]
}
