package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tdmdh/HETFs/internal/ibkr"
	"github.com/tdmdh/HETFs/internal/storage"
)

var (
	dbPath     string
	gatewayURL string

	// ibkrClient is the global IBKR REST client initialized in root
	ibkrClient *ibkr.Client
	// etfRepo is the global local SQLite cache
	etfRepo storage.ETFRepository
)

// rootCmd is the top-level CLI command.
var rootCmd = &cobra.Command{
	Use:   "halalctl",
	Short: "Halal ETF Analyzer — screen ETFs for Shariah compliance via IBKR",
	Long: `halalctl is a CLI tool for screening Exchange-Traded Funds (ETFs)
against AAOIFI Shariah compliance standards using Interactive Brokers.

It interfaces directly with the IBKR Client Portal Gateway.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip setup for help/completion commands.
		if cmd.Name() == "help" || cmd.Name() == "completion" {
			return nil
		}

		// Initialize local cache (SQLite)
		db, err := storage.NewSQLiteDB(dbPath)
		if err != nil {
			return fmt.Errorf("initialise local cache db: %w", err)
		}
		etfRepo = storage.NewETFRepository(db)

		// Initialize IBKR Client, injecting browser session cookie if set.
		// Set SESSION_COOKIE in .env to the raw cookie from your browser after logging in.
		sessionCookie := os.Getenv("SESSION_COOKIE")
		ibkrClient = ibkr.NewClientWithCookie(gatewayURL, sessionCookie)

		// If the command is 'init', we skip the strict auth check and tickler
		// because the 'init' command's job is to start the server!
		if cmd.Name() == "init" {
			return nil
		}

		// Check Auth Status for all other commands
		res, err := ibkrClient.CheckAuthStatus(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to reach gateway at %s: %w", gatewayURL, err)
		}

		if !res.Authenticated {
			return fmt.Errorf("gateway is reachable but NOT authenticated. Please login to IBKR Client Portal Gateway.")
		}

		// Start background tickle routine to keep session alive during long running commands like scanning
		go startTickleRoutine()

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
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "halalctl.db", "path to SQLite cache database")
	rootCmd.PersistentFlags().StringVar(&gatewayURL, "gateway-url", "https://127.0.0.1:5000", "URL of the IBKR Client Portal Gateway")
}

func startTickleRoutine() {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	// Use background context for the keepalive routine so long operations don't get tied to a dying context
	ctx := context.Background()

	for {
		<-ticker.C
		err := ibkrClient.Tickle(ctx)
		if err != nil {
			// A failed tickle might mean the session dropped. For a CLI tool, we just log it or ignore
			// since the main command is probably actively running or will fail soon anyway.
			log.Printf("Warning: gateway tickle failed: %v", err)
		}
	}
}
