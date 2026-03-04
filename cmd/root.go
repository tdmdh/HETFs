package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tdmdh/HETFs/internal/provider/mock"
	"github.com/tdmdh/HETFs/internal/service"
	"github.com/tdmdh/HETFs/internal/storage"
)

var (
	dbPath     string
	etfService *service.ETFService
)

// rootCmd is the top-level CLI command.
var rootCmd = &cobra.Command{
	Use:   "halalctl",
	Short: "Halal ETF Analyzer — screen ETFs for Shariah compliance",
	Long: `halalctl is a CLI tool for screening Exchange-Traded Funds (ETFs)
against AAOIFI Shariah compliance standards.

Phase 1: Fetch and list ETF metadata from supported exchanges.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip setup for help/completion commands.
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		db, err := storage.NewSQLiteDB(dbPath)
		if err != nil {
			return fmt.Errorf("initialise database: %w", err)
		}

		repo := storage.NewETFRepository(db)
		provider := mock.New()
		etfService = service.NewETFService(provider, repo)

		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "halalctl.db", "path to SQLite database file")
}
