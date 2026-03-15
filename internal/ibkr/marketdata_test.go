package ibkr_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/tdmdh/HETFs/internal/ibkr"
)

func TestClient_GetMarketDataSnapshots_Chunking(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)

		conidsStr := r.URL.Query().Get("conids")
		if conidsStr == "" {
			t.Fatal("empty conids parameter")
		}

		conids := strings.Split(conidsStr, ",")
		
		// The chunk limit is 100.
		if len(conids) > 100 {
			t.Fatalf("received batch size %d, which is > 100", len(conids))
		}

		// Mock a response for each requested conid
		respArray := []map[string]interface{}{}
		for _, c := range conids {
			// Basic mock giving each conid a $100 price and 1.5% change
			item := map[string]interface{}{
				"conid": c, // Note: returning string conid to test parsing resilience 
				"55":    "SYM" + c,
				"31":    "C100.00",
				"83":    "1.50%",
			}
			respArray = append(respArray, item)
		}

		w.WriteHeader(http.StatusOK)
		// We can just dump JSON
		fmt.Fprintf(w, "[")
		for i, item := range respArray {
			if i > 0 {
				fmt.Fprintf(w, ",")
			}
			fmt.Fprintf(w, `{"conid":%s, "55":"%s", "31":"%s", "83":"%s"}`, item["conid"], item["55"], item["31"], item["83"])
		}
		fmt.Fprintf(w, "]")
	}))
	defer server.Close()

	client := ibkr.NewClient(server.URL)

	// We create 250 conids, which should result in exactly 3 chunks (100, 100, 50).
	var conids []int64
	for i := 1; i <= 250; i++ {
		conids = append(conids, int64(10000+i))
	}

	snapshots, err := client.GetMarketDataSnapshots(context.Background(), conids)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if atomic.LoadInt32(&requestCount) != 3 {
		t.Fatalf("expected exactly 3 request chunks, got %d", atomic.LoadInt32(&requestCount))
	}

	if len(snapshots) != 250 {
		t.Fatalf("expected exactly 250 parsed snapshots, got %d", len(snapshots))
	}

	// Spot check parsing logic
	if snapshots[0].LastPrice != 100.00 {
		t.Errorf("expected parsed string price 'C100.00' to be 100.0, got %f", snapshots[0].LastPrice)
	}
	if snapshots[0].ChangePercent != 1.5 {
		t.Errorf("expected parsed string change '1.50%%' to be 1.5, got %f", snapshots[0].ChangePercent)
	}
}
