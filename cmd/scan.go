package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var scanLocation string

// scanCmd runs an IBKR market scanner
var scanCmd = &cobra.Command{
	Use:   "scan [scanCode]",
	Short: "Run IBKR market scanner for ETFs",
	Example: `  halalctl scan TOP_PERC_GAIN --location ETF.US.MAJOR`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scanCode := args[0]
		location := scanLocation

		fmt.Printf("Running ETF market scanner '%s' at '%s'...\n", scanCode, location)

		contracts, err := ibkrClient.RunScanner(cmd.Context(), location, scanCode)
		if err != nil {
			return fmt.Errorf("scanner failed: %w", err)
		}

		if len(contracts) == 0 {
			fmt.Println("Scanner returned no results.")
			return nil
		}

		fmt.Printf("Found %d ETFs. Fetching real-time market data...\n", len(contracts))

		var conids []int64
		// map for O(1) matching later
		contractMap := make(map[int64]string)
		for _, c := range contracts {
			conids = append(conids, c.ConID)
			contractMap[c.ConID] = c.CompanyName
		}

		snapshots, err := ibkrClient.GetMarketDataSnapshots(cmd.Context(), conids)
		if err != nil {
			return fmt.Errorf("failed fetching market data snapshots: %w", err)
		}

		// Render table output
		fmt.Println()
		table := tablewriter.NewWriter(os.Stdout)
		table.Header("Symbol", "Company Name", "ConID", "Last Price", "% Change")

		for _, snap := range snapshots {
			compName := contractMap[snap.ConID]
			// format price and change
			priceStr := fmt.Sprintf("$%.2f", snap.LastPrice)
			changeStr := fmt.Sprintf("%.2f%%", snap.ChangePercent)
			if snap.LastPrice == 0 {
				priceStr = "N/A"
			}
			
			table.Append(
				snap.Symbol,
				compName,
				fmt.Sprintf("%d", snap.ConID),
				priceStr,
				changeStr,
			)
		}
		table.Render()

		return nil
	},
}

func init() {
	scanCmd.Flags().StringVar(&scanLocation, "location", "ETF.US.MAJOR", "Scanner location code")
	rootCmd.AddCommand(scanCmd)
}
