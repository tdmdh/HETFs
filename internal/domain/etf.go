package domain

// Contract represents an IBKR tradable instrument (e.g. an ETF).
type Contract struct {
	ID          int64  `json:"id"`
	ConID       int64  `json:"conid"`
	Symbol      string `json:"symbol"`
	CompanyName string `json:"companyName"`
	Exchange    string `json:"exchange"`
}
