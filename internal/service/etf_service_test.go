package service_test

import (
	"context"
	"testing"

	"github.com/tdmdh/HETFs/internal/provider/mock"
	"github.com/tdmdh/HETFs/internal/service"
	"github.com/tdmdh/HETFs/internal/storage"
)

func newTestService(t *testing.T) *service.ETFService {
	t.Helper()
	db, err := storage.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := storage.NewETFRepository(db)
	provider := mock.New()
	return service.NewETFService(provider, repo)
}

func TestFetchAndStore(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	count, err := svc.FetchAndStore(ctx, "LSE")
	if err != nil {
		t.Fatalf("FetchAndStore failed: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected 4 ETFs fetched, got %d", count)
	}

	// Verify they were persisted.
	etfs, err := svc.List(ctx, "LSE")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(etfs) != 4 {
		t.Fatalf("expected 4 ETFs in storage, got %d", len(etfs))
	}
}

func TestFetchAndStore_UnknownExchange(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	count, err := svc.FetchAndStore(ctx, "UNKNOWN")
	if err != nil {
		t.Fatalf("FetchAndStore failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 ETFs for unknown exchange, got %d", count)
	}
}

func TestList_All(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	// Fetch both exchanges.
	svc.FetchAndStore(ctx, "LSE")
	svc.FetchAndStore(ctx, "XETRA")

	etfs, err := svc.List(ctx, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(etfs) != 6 {
		t.Fatalf("expected 6 total ETFs, got %d", len(etfs))
	}
}

func TestList_FilteredByExchange(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	svc.FetchAndStore(ctx, "LSE")
	svc.FetchAndStore(ctx, "XETRA")

	xetra, err := svc.List(ctx, "XETRA")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(xetra) != 2 {
		t.Fatalf("expected 2 XETRA ETFs, got %d", len(xetra))
	}
}

func TestFetchAndStore_Idempotent(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	svc.FetchAndStore(ctx, "LSE")
	svc.FetchAndStore(ctx, "LSE") // second fetch should not duplicate.

	etfs, err := svc.List(ctx, "LSE")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(etfs) != 4 {
		t.Fatalf("expected 4 ETFs after double fetch, got %d", len(etfs))
	}
}
