package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UpsertPage inserts or updates a page. Returns the page ID.
func UpsertPage(ctx context.Context, pool *pgxpool.Pool, p *PageData) (*PageData, error) {
	var d PageData
	err := pool.QueryRow(ctx,
		`INSERT INTO page_data (url, title, meta_description, h1, h2, h3, word_count, image_count, video_count)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 ON CONFLICT (url) DO UPDATE SET
		   title=EXCLUDED.title, meta_description=EXCLUDED.meta_description,
		   h1=EXCLUDED.h1, h2=EXCLUDED.h2, h3=EXCLUDED.h3,
		   word_count=EXCLUDED.word_count, image_count=EXCLUDED.image_count,
		   video_count=EXCLUDED.video_count, updated_at=NOW()
		 RETURNING id, url, title, meta_description, h1, h2, h3, word_count, image_count, video_count, crawled_at, updated_at`,
		p.URL, p.Title, p.MetaDescription, p.H1, p.H2, p.H3, p.WordCount, p.ImageCount, p.VideoCount,
	).Scan(&d.ID, &d.URL, &d.Title, &d.MetaDescription, &d.H1, &d.H2, &d.H3,
		&d.WordCount, &d.ImageCount, &d.VideoCount, &d.CrawledAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert page: %w", err)
	}
	return &d, nil
}

// GetPageByURL returns a page by URL.
func GetPageByURL(ctx context.Context, pool *pgxpool.Pool, url string) (*PageData, error) {
	var d PageData
	err := pool.QueryRow(ctx,
		`SELECT id, url, title, meta_description, h1, h2, h3, word_count, image_count, video_count, crawled_at, updated_at
		 FROM page_data WHERE url=$1`, url,
	).Scan(&d.ID, &d.URL, &d.Title, &d.MetaDescription, &d.H1, &d.H2, &d.H3,
		&d.WordCount, &d.ImageCount, &d.VideoCount, &d.CrawledAt, &d.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get page by url: %w", err)
	}
	return &d, nil
}

// ── Keyword-Page Relations (Inverted Index) ───────────────────

// InsertKeywordPage relates a keyword to a page with TF-IDF data.
func InsertKeywordPage(ctx context.Context, pool *pgxpool.Pool, kp *KeywordPage) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO keyword_pages (keyword_id, page_id, position, tfidf_score, in_heading)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (keyword_id, page_id) DO UPDATE SET
		   position=EXCLUDED.position, tfidf_score=EXCLUDED.tfidf_score, in_heading=EXCLUDED.in_heading`,
		kp.KeywordID, kp.PageID, kp.Position, kp.TFIDFScore, kp.InHeading,
	)
	if err != nil {
		return fmt.Errorf("insert keyword page: %w", err)
	}
	return nil
}

// GetPagesByKeyword returns pages that rank for a keyword (forward index).
func GetPagesByKeyword(ctx context.Context, pool *pgxpool.Pool, keywordID string, limit int) ([]KeywordPage, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := pool.Query(ctx,
		`SELECT keyword_id, page_id, position, tfidf_score, in_heading, created_at
		 FROM keyword_pages WHERE keyword_id=$1
		 ORDER BY position ASC LIMIT $2`,
		keywordID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get pages by keyword: %w", err)
	}
	defer rows.Close()

	var kps []KeywordPage
	for rows.Next() {
		var kp KeywordPage
		if err := rows.Scan(&kp.KeywordID, &kp.PageID, &kp.Position, &kp.TFIDFScore, &kp.InHeading, &kp.CreatedAt); err != nil {
			return nil, fmt.Errorf("get pages by keyword: scan: %w", err)
		}
		kps = append(kps, kp)
	}
	return kps, rows.Err()
}

// GetKeywordsByPage returns keywords a page ranks for (reverse index).
func GetKeywordsByPage(ctx context.Context, pool *pgxpool.Pool, pageID string, limit int) ([]KeywordPage, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := pool.Query(ctx,
		`SELECT keyword_id, page_id, position, tfidf_score, in_heading, created_at
		 FROM keyword_pages WHERE page_id=$1
		 ORDER BY position ASC LIMIT $2`,
		pageID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get keywords by page: %w", err)
	}
	defer rows.Close()

	var kps []KeywordPage
	for rows.Next() {
		var kp KeywordPage
		if err := rows.Scan(&kp.KeywordID, &kp.PageID, &kp.Position, &kp.TFIDFScore, &kp.InHeading, &kp.CreatedAt); err != nil {
			return nil, fmt.Errorf("get keywords by page: scan: %w", err)
		}
		kps = append(kps, kp)
	}
	return kps, rows.Err()
}
