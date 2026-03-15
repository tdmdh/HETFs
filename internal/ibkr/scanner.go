package ibkr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tdmdh/HETFs/internal/domain"
)

// ScannerPayload defines the parameters for a market scan.
type ScannerPayload struct {
	Instrument      string `json:"instrument"`
	Type            string `json:"type"` // e.g., "MKT_CAP_USD_DESC" normally, but varies
	Location        string `json:"location"` // e.g., "ETF.US.MAJOR"
	Filter          []Filter `json:"filter,omitempty"`
}

type Filter struct {
	Code  string `json:"code"`
	Value string `json:"value"`
}

// ScannerResult item
type ScannerResult struct {
	ConID           int64   `json:"conid"`
	Symbol          string  `json:"symbol"`
	CompanyName     string  `json:"companyName"`
	LastPrice       float64 `json:"lastPrice,omitempty"` // Sometimes scanner returns this directly
	ChangePercent   float64 `json:"changePercent,omitempty"`
}

// RunScanner hits the IBKR scanner endpoint to discover ETFs.
func (c *Client) RunScanner(ctx context.Context, location string, scanCode string) ([]domain.Contract, error) {
	// The IBKR scanner payload can have either "instrument":"ETF" or 
	// specific fields depending on the exact scan parameters.
	payload := map[string]interface{}{
		"instrument": "ETF",
		"location":   location,
		"type":       scanCode, // "type" acts as the scanCode, e.g., "HIGH_OPT_VOLUME_PUT_CALL_RATIO"
	}

	resp, err := c.Post(ctx, "/v1/api/iserver/scanner/run", payload)
	if err != nil {
		return nil, fmt.Errorf("scanner run request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("scanner unexpected status %d: %s", resp.StatusCode, string(b))
	}

	var parsed []ScannerResult
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode scanner response: %w", err)
	}

	var results []domain.Contract
	for _, p := range parsed {
		results = append(results, domain.Contract{
			ConID:       p.ConID,
			Symbol:      p.Symbol,
			CompanyName: p.CompanyName,
		})
	}

	return results, nil
}
