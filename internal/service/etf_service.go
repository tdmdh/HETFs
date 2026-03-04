package service

import (
	"context"
	"fmt"

	"github.com/tdmdh/HETFs/internal/domain"
	"github.com/tdmdh/HETFs/internal/provider"
	"github.com/tdmdh/HETFs/internal/storage"
)

// ETFService orchestrates fetching ETFs from a provider and persisting them.
type ETFService struct {
	provider provider.ETFProvider
	repo     storage.ETFRepository
}

// NewETFService creates a new ETFService.
func NewETFService(p provider.ETFProvider, r storage.ETFRepository) *ETFService {
	return &ETFService{provider: p, repo: r}
}

// FetchAndStore fetches ETFs for the given exchange from the provider and
// upserts them into storage. Returns the number of ETFs stored.
func (s *ETFService) FetchAndStore(ctx context.Context, exchange string) (int, error) {
	etfs, err := s.provider.FetchETFs(ctx, exchange)
	if err != nil {
		return 0, fmt.Errorf("fetch etfs from provider: %w", err)
	}

	if len(etfs) == 0 {
		return 0, nil
	}

	if err := s.repo.UpsertETFs(ctx, etfs); err != nil {
		return 0, fmt.Errorf("persist etfs: %w", err)
	}

	return len(etfs), nil
}

// List returns stored ETFs. If exchange is empty, all ETFs are returned.
func (s *ETFService) List(ctx context.Context, exchange string) ([]domain.ETF, error) {
	if exchange == "" {
		return s.repo.ListAllETFs(ctx)
	}
	return s.repo.ListETFs(ctx, exchange)
}
