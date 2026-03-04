package mock_test

import (
	"context"
	"testing"

	"github.com/tdmdh/HETFs/internal/provider/mock"
)

func TestFetchETFs_LSE(t *testing.T) {
	p := mock.New()
	etfs, err := p.FetchETFs(context.Background(), "LSE")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(etfs) != 4 {
		t.Fatalf("expected 4 LSE ETFs, got %d", len(etfs))
	}

	// Verify all returned ETFs belong to LSE.
	for _, e := range etfs {
		if e.Exchange != "LSE" {
			t.Errorf("expected exchange LSE, got %s for %s", e.Exchange, e.Symbol)
		}
	}
}

func TestFetchETFs_XETRA(t *testing.T) {
	p := mock.New()
	etfs, err := p.FetchETFs(context.Background(), "XETRA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(etfs) != 2 {
		t.Fatalf("expected 2 XETRA ETFs, got %d", len(etfs))
	}
	for _, e := range etfs {
		if e.Exchange != "XETRA" {
			t.Errorf("expected exchange XETRA, got %s for %s", e.Exchange, e.Symbol)
		}
	}
}

func TestFetchETFs_CaseInsensitive(t *testing.T) {
	p := mock.New()
	etfs, err := p.FetchETFs(context.Background(), "lse")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(etfs) != 4 {
		t.Fatalf("expected 4 LSE ETFs for lowercase 'lse', got %d", len(etfs))
	}
}

func TestFetchETFs_UnknownExchange(t *testing.T) {
	p := mock.New()
	etfs, err := p.FetchETFs(context.Background(), "UNKNOWN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(etfs) != 0 {
		t.Fatalf("expected 0 ETFs for unknown exchange, got %d", len(etfs))
	}
}

func TestFetchETFs_ReturnsCopy(t *testing.T) {
	p := mock.New()

	etfs1, _ := p.FetchETFs(context.Background(), "LSE")
	etfs1[0].Symbol = "MUTATED"

	etfs2, _ := p.FetchETFs(context.Background(), "LSE")
	if etfs2[0].Symbol == "MUTATED" {
		t.Fatal("FetchETFs returned the original slice instead of a copy")
	}
}
