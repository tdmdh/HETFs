package ibkr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tdmdh/HETFs/internal/domain"
)

// SecDefSearchParam is the request payload for secdef search.
type SecDefSearchParam struct {
	Symbol      string `json:"symbol"`
	SecType     string `json:"secType"` // usually "STK" for ETFs
	Name        bool   `json:"name"`
}

// SecDefSearchResp maps the response item from the search endpoint.
type SecDefSearchResp struct {
	ConID       int64  `json:"conid"`
	CompanyHex  string `json:"companyHex"`
	CompanyName string `json:"companyName"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Exchange    string `json:"exchange"` // Note: often missing or generic for search
}

// SearchContract finds instrument mappings (conid) by symbol.
func (c *Client) SearchContract(ctx context.Context, symbol string) ([]domain.Contract, error) {
	payload := SecDefSearchParam{
		Symbol:  symbol,
		Name:    false,
		SecType: "STK", // ETFs are queried as stocks first
	}

	resp, err := c.Post(ctx, "/v1/api/iserver/secdef/search", payload)
	if err != nil {
		return nil, fmt.Errorf("secdef search req: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
	}

	var results []SecDefSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	var contracts []domain.Contract
	for _, r := range results {
		// Example filter: we want the main US venues for these ETFs typically.
		// The `search` API responses are messy. For robust searching we take ARCA or SMART.
		// Sometimes exchange isn't populated depending on exact response format,
		// but we map it over anyway.
		contracts = append(contracts, domain.Contract{
			ConID:       r.ConID,
			Symbol:      r.Symbol,
			CompanyName: r.CompanyName,
			Exchange:    r.Exchange, // might be empty or "SMART"
		})
	}
	return contracts, nil
}
