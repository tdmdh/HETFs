package domain

// ETF represents a single Exchange-Traded Fund listing.
type ETF struct {
	ID       int64  `json:"id"`
	Symbol   string `json:"symbol"`   // e.g. "ISWD"
	ISIN     string `json:"isin"`     // e.g. "IE00B27YCN58"
	Exchange string `json:"exchange"` // e.g. "LSE"
	Currency string `json:"currency"` // e.g. "GBP"
}
