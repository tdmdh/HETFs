package ibkr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// MarketSnapshot represents price and change details.
type MarketSnapshot struct {
	ConID           int64
	Symbol          string
	LastPrice       float64
	ChangePercent   float64
	CompanyName     string // Added locally from contract if missing
}

// GetMarketDataSnapshots fetches batch prices for up to 100 conids per request.
func (c *Client) GetMarketDataSnapshots(ctx context.Context, conids []int64) ([]MarketSnapshot, error) {
	if len(conids) == 0 {
		return nil, nil
	}

	var allSnapshots []MarketSnapshot

	// Chunk conids into batches of 100
	const batchSize = 100
	for i := 0; i < len(conids); i += batchSize {
		end := i + batchSize
		if end > len(conids) {
			end = len(conids)
		}
		batch := conids[i:end]

		// Format "conid1,conid2,conid3"
		strConids := make([]string, len(batch))
		for j, v := range batch {
			strConids[j] = strconv.FormatInt(v, 10)
		}
		conidList := strings.Join(strConids, ",")

		// Query params
		// Field 31 = last price, 83 = change percent, 55 = symbol
		endpoint := fmt.Sprintf("/v1/api/iserver/marketdata/snapshot?conids=%s&fields=31,55,83", conidList)

		resp, err := c.Get(ctx, endpoint)
		if err != nil {
			return nil, fmt.Errorf("snapshot request: %w", err)
		}

		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("snapshot unexpected status %d: %s", resp.StatusCode, string(b))
		}

		var chunkResults []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&chunkResults); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode snapshot response: %w", err)
		}
		resp.Body.Close()

		for _, item := range chunkResults {
			var snap MarketSnapshot
			if conidFloat, ok := item["conid"].(float64); ok {
				snap.ConID = int64(conidFloat)
			}
			if conidInt, ok := item["conid"].(int); ok {
				snap.ConID = int64(conidInt)
			}
			
			if priceFloat, ok := item["31"].(float64); ok {
				var priceStr string
				if s, ok := item["31"].(string); ok {
					priceStr = s
					// remove char prefixes sometimes IBKR uses
					priceStr = strings.TrimPrefix(priceStr, "C")
					if p, err := strconv.ParseFloat(priceStr, 64); err == nil {
						snap.LastPrice = p
					}
				} else {
					snap.LastPrice = priceFloat
				}
			} else if priceStr, ok := item["31"].(string); ok {
				priceStr = strings.TrimPrefix(priceStr, "C")
				if p, err := strconv.ParseFloat(priceStr, 64); err == nil {
					snap.LastPrice = p
				}
			}

			if changeStr, ok := item["83"].(string); ok {
				changeStr = strings.TrimSuffix(changeStr, "%")
				if val, err := strconv.ParseFloat(changeStr, 64); err == nil {
					snap.ChangePercent = val
				}
			} else if changeFloat, ok := item["83"].(float64); ok {
				snap.ChangePercent = changeFloat
			}

			if symbol, ok := item["55"].(string); ok {
				snap.Symbol = symbol
			}

			allSnapshots = append(allSnapshots, snap)
		}
	}

	return allSnapshots, nil
}
