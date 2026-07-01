package storage

import "time"

// ── Projects ──────────────────────────────────────────────────

type Project struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	BusinessDesc   string    `json:"business_desc,omitempty"`
	TargetAudience string    `json:"target_audience,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ── Keywords ──────────────────────────────────────────────────

type SeedKeyword struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Keyword   string    `json:"keyword"`
	CreatedAt time.Time `json:"created_at"`
}

type Keyword struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Keyword     string    `json:"keyword"`
	IsSeed      bool      `json:"is_seed"`
	ClusterID   *string   `json:"cluster_id,omitempty"`
	ClusterName *string   `json:"cluster_name,omitempty"`
	Intent      *string   `json:"intent,omitempty"`
	Source      *string   `json:"source,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type KeywordMetric struct {
	KeywordID   string    `json:"keyword_id"`
	Timestamp   time.Time `json:"timestamp"`
	Volume      *int      `json:"volume,omitempty"`
	CPCCents    *int      `json:"cpc_cents,omitempty"`
	Competition *int      `json:"competition,omitempty"`
	HeatScore   *float64  `json:"heat_score,omitempty"`
	SerpPosition *int     `json:"serp_position,omitempty"`
}

// ── Clusters ──────────────────────────────────────────────────

type Cluster struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Name         string    `json:"name"`
	Intent       *string   `json:"intent,omitempty"`
	KeywordCount int       `json:"keyword_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ── SERP ──────────────────────────────────────────────────────

type SERPResult struct {
	ID        string    `json:"id"`
	KeywordID string    `json:"keyword_id"`
	Position  int       `json:"position"`
	URL       string    `json:"url"`
	Title     string    `json:"title,omitempty"`
	Snippet   string    `json:"snippet,omitempty"`
	CrawledAt time.Time `json:"crawled_at"`
}

// ── Pages ─────────────────────────────────────────────────────

type PageData struct {
	ID              string    `json:"id"`
	URL             string    `json:"url"`
	Title           string    `json:"title,omitempty"`
	MetaDescription string    `json:"meta_description,omitempty"`
	H1              []string  `json:"h1,omitempty"`
	H2              []string  `json:"h2,omitempty"`
	H3              []string  `json:"h3,omitempty"`
	WordCount       int       `json:"word_count"`
	ImageCount      int       `json:"image_count"`
	VideoCount      int       `json:"video_count"`
	CrawledAt       time.Time `json:"crawled_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type KeywordPage struct {
	KeywordID  string    `json:"keyword_id"`
	PageID     string    `json:"page_id"`
	Position   *int      `json:"position,omitempty"`
	TFIDFScore *float64  `json:"tfidf_score,omitempty"`
	InHeading  bool      `json:"in_heading"`
	CreatedAt  time.Time `json:"created_at"`
}

// ── Jobs ──────────────────────────────────────────────────────

type Job struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Payload   []byte    `json:"payload,omitempty"`
	Result    []byte    `json:"result,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JobStatus string

const (
	JobPending    JobStatus = "PENDING"
	JobProcessing JobStatus = "PROCESSING"
	JobCompleted  JobStatus = "COMPLETED"
	JobFailed     JobStatus = "FAILED"
)

// ── Query helpers ─────────────────────────────────────────────

type ListOpts struct {
	Limit  int
	Offset int
}

// ── Content Gaps ──────────────────────────────────────────────

type ContentGap struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	KeywordID string    `json:"keyword_id"`
	Keyword   string    `json:"keyword"`
	Gaps      string    `json:"gaps"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Competitors ─────────────────────────────────────────────

type Competitor struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
