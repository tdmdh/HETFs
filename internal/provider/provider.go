package provider

import (
	"context"

	"github.com/tdmdh/HETFs/internal/domain"
)

// ETFProvider defines the contract for fetching ETF metadata from any source.
// Implementations may talk to real APIs (IBKR, EODHD) or return mock data.
type ETFProvider interface {
	// FetchETFs returns ETFs listed on the given exchange.
	// Returns an empty slice (not an error) if the exchange is not recognised.
	FetchETFs(ctx context.Context, exchange string) ([]domain.ETF, error)
}
