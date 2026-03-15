package ibkr_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tdmdh/HETFs/internal/ibkr"
)

func TestClient_RateLimiter(t *testing.T) {
	// A simple mock server that tracks requests
	var reqCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := ibkr.NewClient(server.URL)

	start := time.Now()
	// Send 15 requests as fast as possible
	for i := 0; i < 15; i++ {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		_, err := client.Do(req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	// Since limit is ~9 req/sec and burst is 1, 15 requests should take at least ~1.5 seconds.
	if elapsed < 1*time.Second {
		t.Fatalf("rate limiter did not strictly limit requests, elapsed: %v", elapsed)
	}

	if atomic.LoadInt32(&reqCount) != 15 {
		t.Fatalf("expected 15 requests, got %d", atomic.LoadInt32(&reqCount))
	}
}

func TestClient_RetryOn429(t *testing.T) {
	var reqCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&reqCount, 1)
		if count <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := ibkr.NewClient(server.URL)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	
	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// It should have failed twice, meaning it waited 1s then 2s = ~3s.
	if atomic.LoadInt32(&reqCount) != 3 {
		t.Fatalf("expected 3 attempts, got %d", atomic.LoadInt32(&reqCount))
	}
	
	if elapsed < 3*time.Second {
		t.Fatalf("expected backoff delay ~3s, got %v", elapsed)
	}
}
