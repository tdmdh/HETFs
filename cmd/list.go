package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listExchange string

// listCmd is the parent "list" command group.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored data",
}

// listETFsCmd lists stored ETFs, optionally filtered by exchange.
var listETFsCmd = &cobra.Command{
	Use:   "etfs",
	Short: "List stored ETF listings",
	Example: `  halalctl list etfs
  halalctl list etfs --exchange LSE`,
	RunE: func(cmd *cobra.Command, args []string) error {
		exchange := strings.ToUpper(listExchange)

		etfs, err := etfService.List(cmd.Context(), exchange)
		if err != nil {
			return fmt.Errorf("list etfs: %w", err)
		}

		if len(etfs) == 0 {
			if exchange != "" {
				fmt.Printf("No ETFs found for exchange %s. Have you run 'halalctl fetch etfs --exchange %s'?\n", exchange, exchange)
			} else {
				fmt.Println("No ETFs found. Run 'halalctl fetch etfs --exchange <EXCHANGE>' first.")
			}
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SYMBOL\tISIN\tEXCHANGE\tCURRENCY")
		for _, e := range etfs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Symbol, e.ISIN, e.Exchange, e.Currency)
		}
		w.Flush()

		return nil
	},
}

func init() {
	listETFsCmd.Flags().StringVar(&listExchange, "exchange", "", "filter by exchange (optional)")

	listCmd.AddCommand(listETFsCmd)
	rootCmd.AddCommand(listCmd)
}
