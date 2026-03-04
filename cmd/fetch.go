package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var fetchExchange string

// fetchCmd is the parent "fetch" command group.
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch data from providers",
}

// fetchETFsCmd fetches ETFs for a given exchange and stores them.
var fetchETFsCmd = &cobra.Command{
	Use:   "etfs",
	Short: "Fetch ETF listings for an exchange",
	Example: `  halalctl fetch etfs --exchange LSE
  halalctl fetch etfs --exchange XETRA`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exchange := strings.ToUpper(fetchExchange)

		count, err := etfService.FetchAndStore(cmd.Context(), exchange)
		if err != nil {
			return fmt.Errorf("fetch etfs: %w", err)
		}

		if count == 0 {
			fmt.Printf("No ETFs found for exchange %s\n", exchange)
			return nil
		}

		fmt.Printf("Fetched %d ETFs for exchange %s\n", count, exchange)
		return nil
	},
}

func init() {
	fetchETFsCmd.Flags().StringVar(&fetchExchange, "exchange", "", "exchange to fetch ETFs from (e.g. LSE, XETRA)")
	fetchETFsCmd.MarkFlagRequired("exchange") //nolint:errcheck

	fetchCmd.AddCommand(fetchETFsCmd)
	rootCmd.AddCommand(fetchCmd)
}
