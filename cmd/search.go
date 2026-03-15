package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tdmdh/HETFs/internal/domain"
)

// searchCmd searches for an ETF by symbol to find its conid.
var searchCmd = &cobra.Command{
	Use:   "search [symbol]",
	Short: "Search IBKR for an ETF by symbol to discover its conid",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol := strings.ToUpper(args[0])

		fmt.Printf("Searching for '%s'...\n", symbol)
		contracts, err := ibkrClient.SearchContract(cmd.Context(), symbol)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if len(contracts) == 0 {
			fmt.Println("No matches found.")
			return nil
		}

		// Save the first robust match to the local cache.
		// For simplicity, we just pick the first one and upsert it.
		var bestMatch domain.Contract
		for _, c := range contracts {
			// We prefer ARCA or SMART, or just take the first since IBKR often returns best match first
			bestMatch = c
			break
		}

		fmt.Printf("Found Match: %s (conid: %d, Exchange: %s)\n", bestMatch.CompanyName, bestMatch.ConID, bestMatch.Exchange)

		err = etfRepo.UpsertContract(cmd.Context(), bestMatch)
		if err != nil {
			return fmt.Errorf("failed to save contract to local cache: %w", err)
		}
		fmt.Printf("Saved %s to local cache.\n", symbol)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
