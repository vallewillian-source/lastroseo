package storage

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BatchWriter buffers metric records and flushes them via pgx COPY protocol.
// 100x throughput vs row-by-row INSERT.
type BatchWriter struct {
	mu       sync.Mutex
	pool     *pgxpool.Pool
	buffer   []KeywordMetric
	maxSize  int
	interval time.Duration
	ticker   *time.Ticker
	done     chan struct{}
	closed   bool
}

// NewBatchWriter creates a BatchWriter that flushes every maxSize records or interval.
func NewBatchWriter(pool *pgxpool.Pool, maxSize int, interval time.Duration) *BatchWriter {
	bw := &BatchWriter{
		pool:     pool,
		buffer:   make([]KeywordMetric, 0, maxSize),
		maxSize:  maxSize,
		interval: interval,
		ticker:   time.NewTicker(interval),
		done:     make(chan struct{}),
	}
	go bw.loop()
	return bw
}

// Add enqueues a KeywordMetric for batch writing.
func (bw *BatchWriter) Add(m KeywordMetric) {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	if bw.closed {
		return
	}
	bw.buffer = append(bw.buffer, m)
	if len(bw.buffer) >= bw.maxSize {
		bw.flushLocked()
	}
}

// Close stops the flush loop and writes remaining records.
func (bw *BatchWriter) Close() {
	bw.mu.Lock()
	bw.closed = true
	bw.mu.Unlock()

	bw.ticker.Stop()
	close(bw.done)

	bw.mu.Lock()
	bw.flushLocked()
	bw.mu.Unlock()
}

func (bw *BatchWriter) loop() {
	for {
		select {
		case <-bw.ticker.C:
			bw.mu.Lock()
			bw.flushLocked()
			bw.mu.Unlock()
		case <-bw.done:
			return
		}
	}
}

func (bw *BatchWriter) flushLocked() {
	if len(bw.buffer) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use pgx COPY for maximum throughput
	rows := make([][]interface{}, len(bw.buffer))
	for i, m := range bw.buffer {
		rows[i] = []interface{}{
			m.KeywordID, m.Timestamp, m.Volume, m.CPCCents, m.Competition, m.HeatScore, m.SerpPosition,
		}
	}

	_, err := bw.pool.CopyFrom(
		ctx,
		pgx.Identifier{"keyword_metrics"},
		[]string{"keyword_id", "timestamp", "volume", "cpc_cents", "competition", "heat_score", "serp_position"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		log.Printf("batch writer: flush error: %v (retrying later)", err)
		// Keep buffer on failure — will retry next cycle
		return
	}

	log.Printf("batch writer: flushed %d metric rows via COPY", len(bw.buffer))
	bw.buffer = bw.buffer[:0]
}

// ── SERP Batch Writer ─────────────────────────────────────────

// SERPBatchWriter buffers and flushes SERP results via COPY.
type SERPBatchWriter struct {
	mu       sync.Mutex
	pool     *pgxpool.Pool
	buffer   []SERPResult
	maxSize  int
	interval time.Duration
	ticker   *time.Ticker
	done     chan struct{}
	closed   bool
}

// NewSERPBatchWriter creates a batch writer for SERP results.
func NewSERPBatchWriter(pool *pgxpool.Pool, maxSize int, interval time.Duration) *SERPBatchWriter {
	bw := &SERPBatchWriter{
		pool:     pool,
		buffer:   make([]SERPResult, 0, maxSize),
		maxSize:  maxSize,
		interval: interval,
		ticker:   time.NewTicker(interval),
		done:     make(chan struct{}),
	}
	go bw.loop()
	return bw
}

// Add enqueues a SERPResult for batch writing.
func (bw *SERPBatchWriter) Add(r SERPResult) {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	if bw.closed {
		return
	}
	bw.buffer = append(bw.buffer, r)
	if len(bw.buffer) >= bw.maxSize {
		bw.flushLocked()
	}
}

// Close stops the flush loop and writes remaining records.
func (bw *SERPBatchWriter) Close() {
	bw.mu.Lock()
	bw.closed = true
	bw.mu.Unlock()
	bw.ticker.Stop()
	close(bw.done)
	bw.mu.Lock()
	bw.flushLocked()
	bw.mu.Unlock()
}

func (bw *SERPBatchWriter) loop() {
	for {
		select {
		case <-bw.ticker.C:
			bw.mu.Lock()
			bw.flushLocked()
			bw.mu.Unlock()
		case <-bw.done:
			return
		}
	}
}

func (bw *SERPBatchWriter) flushLocked() {
	if len(bw.buffer) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows := make([][]interface{}, len(bw.buffer))
	for i, r := range bw.buffer {
		rows[i] = []interface{}{
			r.KeywordID, r.Position, r.URL, r.Title, r.Snippet, r.CrawledAt,
		}
	}

	_, err := bw.pool.CopyFrom(
		ctx,
		pgx.Identifier{"serp_results"},
		[]string{"keyword_id", "position", "url", "title", "snippet", "crawled_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		log.Printf("serp batch writer: flush error: %v", err)
		return
	}
	log.Printf("serp batch writer: flushed %d SERP rows via COPY", len(bw.buffer))
	bw.buffer = bw.buffer[:0]
}
