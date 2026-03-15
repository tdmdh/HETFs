package ibkr_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tdmdh/HETFs/internal/ibkr"
)

func TestClient_RunScanner(t *testing.T) {
	mockResponse := `[
		{"conid": 12345, "symbol": "ISWD", "companyName": "IShares World Islamic"},
		{"conid": 67890, "symbol": "SWDA", "companyName": "IShares Core World"}
	]`

	var receivedPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/api/iserver/scanner/run" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := ibkr.NewClient(server.URL)

	results, err := client.RunScanner(context.Background(), "ETF.US.MAJOR", "TOP_PERC_GAIN")
	if err != nil {
		t.Fatalf("RunScanner failed: %v", err)
	}

	// Verify the correct payload was sent
	if receivedPayload["instrument"] != "ETF" {
		t.Errorf("expected instrument: ETF, got %v", receivedPayload["instrument"])
	}
	if receivedPayload["location"] != "ETF.US.MAJOR" {
		t.Errorf("expected location: ETF.US.MAJOR, got %v", receivedPayload["location"])
	}
	if receivedPayload["type"] != "TOP_PERC_GAIN" {
		t.Errorf("expected type: TOP_PERC_GAIN, got %v", receivedPayload["type"])
	}

	// Verify results parsing
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ConID != 12345 || results[0].Symbol != "ISWD" {
		t.Errorf("unexpected first result: %+v", results[0])
	}
}
