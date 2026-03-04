package mock

import (
	"context"
	"strings"

	"github.com/tdmdh/HETFs/internal/domain"
)

// etfData holds the hardcoded ETF dataset keyed by exchange (uppercase).
var etfData = map[string][]domain.ETF{
	"LSE": {
		{Symbol: "ISWD", ISIN: "IE00B27YCN58", Exchange: "LSE", Currency: "GBP"},
		{Symbol: "SWDA", ISIN: "IE00B4L5Y983", Exchange: "LSE", Currency: "GBP"},
		{Symbol: "VWRL", ISIN: "IE00B3RBWM25", Exchange: "LSE", Currency: "GBP"},
		{Symbol: "EIMI", ISIN: "IE00BKM4GZ66", Exchange: "LSE", Currency: "GBP"},
	},
	"XETRA": {
		{Symbol: "IUSE", ISIN: "IE00B52SFT06", Exchange: "XETRA", Currency: "EUR"},
		{Symbol: "SXRV", ISIN: "IE00B4L5Y983", Exchange: "XETRA", Currency: "EUR"},
	},
}

// Provider is a mock ETFProvider that returns hardcoded ETF data.
type Provider struct{}

// New returns a new mock Provider.
func New() *Provider {
	return &Provider{}
}

// FetchETFs returns ETFs for the given exchange from the hardcoded dataset.
// Unknown exchanges return an empty slice with no error.
func (p *Provider) FetchETFs(_ context.Context, exchange string) ([]domain.ETF, error) {
	exchange = strings.ToUpper(exchange)
	if etfs, ok := etfData[exchange]; ok {
		// Return a copy to prevent callers from mutating the source data.
		result := make([]domain.ETF, len(etfs))
		copy(result, etfs)
		return result, nil
	}
	return []domain.ETF{}, nil
}
