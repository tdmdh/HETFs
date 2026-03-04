package storage_test

import (
	"context"
	"testing"

	"github.com/tdmdh/HETFs/internal/domain"
	"github.com/tdmdh/HETFs/internal/storage"
)

func newRepo(t *testing.T) storage.ETFRepository {
	t.Helper()
	db, err := storage.NewSQLiteDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return storage.NewETFRepository(db)
}

func TestUpsertAndListAll(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	etf := domain.ETF{Symbol: "ISWD", ISIN: "IE00B27YCN58", Exchange: "LSE", Currency: "GBP"}
	if err := repo.UpsertETF(ctx, etf); err != nil {
		t.Fatalf("UpsertETF failed: %v", err)
	}

	etfs, err := repo.ListAllETFs(ctx)
	if err != nil {
		t.Fatalf("ListAllETFs failed: %v", err)
	}
	if len(etfs) != 1 {
		t.Fatalf("expected 1 ETF, got %d", len(etfs))
	}
	if etfs[0].Symbol != "ISWD" {
		t.Errorf("expected symbol ISWD, got %s", etfs[0].Symbol)
	}
}

func TestUpsertETFs_BulkInsert(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	etfs := []domain.ETF{
		{Symbol: "ISWD", ISIN: "IE00B27YCN58", Exchange: "LSE", Currency: "GBP"},
		{Symbol: "SWDA", ISIN: "IE00B4L5Y983", Exchange: "LSE", Currency: "GBP"},
		{Symbol: "VWRL", ISIN: "IE00B3RBWM25", Exchange: "LSE", Currency: "GBP"},
	}

	if err := repo.UpsertETFs(ctx, etfs); err != nil {
		t.Fatalf("UpsertETFs failed: %v", err)
	}

	result, err := repo.ListAllETFs(ctx)
	if err != nil {
		t.Fatalf("ListAllETFs failed: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 ETFs, got %d", len(result))
	}
}

func TestListETFs_FilterByExchange(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	etfs := []domain.ETF{
		{Symbol: "ISWD", ISIN: "IE00B27YCN58", Exchange: "LSE", Currency: "GBP"},
		{Symbol: "IUSE", ISIN: "IE00B52SFT06", Exchange: "XETRA", Currency: "EUR"},
	}
	if err := repo.UpsertETFs(ctx, etfs); err != nil {
		t.Fatalf("UpsertETFs failed: %v", err)
	}

	lseETFs, err := repo.ListETFs(ctx, "LSE")
	if err != nil {
		t.Fatalf("ListETFs failed: %v", err)
	}
	if len(lseETFs) != 1 {
		t.Fatalf("expected 1 LSE ETF, got %d", len(lseETFs))
	}
	if lseETFs[0].Symbol != "ISWD" {
		t.Errorf("expected ISWD, got %s", lseETFs[0].Symbol)
	}
}

func TestUpsert_Idempotent(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	etf := domain.ETF{Symbol: "ISWD", ISIN: "IE00B27YCN58", Exchange: "LSE", Currency: "GBP"}

	if err := repo.UpsertETF(ctx, etf); err != nil {
		t.Fatalf("first UpsertETF failed: %v", err)
	}
	if err := repo.UpsertETF(ctx, etf); err != nil {
		t.Fatalf("second UpsertETF failed: %v", err)
	}

	etfs, err := repo.ListAllETFs(ctx)
	if err != nil {
		t.Fatalf("ListAllETFs failed: %v", err)
	}
	if len(etfs) != 1 {
		t.Fatalf("expected 1 ETF after double upsert, got %d", len(etfs))
	}
}

func TestUpsert_UpdatesExistingFields(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	etf := domain.ETF{Symbol: "OLD", ISIN: "IE00B27YCN58", Exchange: "LSE", Currency: "GBP"}
	if err := repo.UpsertETF(ctx, etf); err != nil {
		t.Fatalf("UpsertETF failed: %v", err)
	}

	etf.Symbol = "ISWD"
	if err := repo.UpsertETF(ctx, etf); err != nil {
		t.Fatalf("UpsertETF (update) failed: %v", err)
	}

	etfs, err := repo.ListAllETFs(ctx)
	if err != nil {
		t.Fatalf("ListAllETFs failed: %v", err)
	}
	if len(etfs) != 1 {
		t.Fatalf("expected 1 ETF, got %d", len(etfs))
	}
	if etfs[0].Symbol != "ISWD" {
		t.Errorf("expected symbol to be updated to ISWD, got %s", etfs[0].Symbol)
	}
}
