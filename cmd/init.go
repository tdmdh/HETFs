package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Starts the IBKR Gateway via Docker and verifies connectivity",
	Long: `Starts the ghcr.io/rsjethani/ibkr-gateway Docker container
using docker-compose, waits for it to boot, and verifies the /auth/status endpoint.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🚀 Starting IBKR Gateway Docker container...")

		// Run docker-compose up -d
		upCmd := exec.Command("docker-compose", "up", "-d")
		upCmd.Stdout = os.Stdout
		upCmd.Stderr = os.Stderr
		if err := upCmd.Run(); err != nil {
			return fmt.Errorf("failed to start docker container: %w", err)
		}

		fmt.Println("⏳ Waiting for Gateway to boot (this may take 15-30 seconds)...")

		// Poll for gateway readiness
		maxRetries := 30
		delay := 2 * time.Second
		
		var authSuccess bool
		var authMessage string
		
		for i := 1; i <= maxRetries; i++ {
			time.Sleep(delay)
			
			// We use the CheckAuthStatus method. 
			// If it returns an error, the server isn't quite up yet or is refusing connections
			status, err := ibkrClient.CheckAuthStatus(cmd.Context())
			
			if err == nil {
				fmt.Println("\n✅ Gateway is reachable and responding!")
				
				if status.Authenticated {
					fmt.Println("✅ Gateway is AUTHENTICATED. Ready to serve requests.")
					authSuccess = true
				} else {
					fmt.Println("⚠️  Gateway is NOT authenticated yet.")
					fmt.Printf("   Please visit: %s to log in to Interactive Brokers.\n", gatewayURL)
				}
				break
			}
			
			// Print a dot to show progress
			fmt.Print(".")
			authMessage = err.Error()
		}

		if !authSuccess {
			fmt.Printf("\n\nCould not verify authentication within the timeout period.\n")
			fmt.Printf("Last connection error: %s\n", authMessage)
			fmt.Printf("\nPlease ensure you navigate to %s in your browser and log in.\n", gatewayURL)
		}

		return nil
	},
}

func init() {
	// Add initCmd to root.
	// But note: root's PersistentPreRunE checks auth status and fails if not authenticated.
	// For 'init' specifically, we need to bypass this check so we can actually run the command.
	rootCmd.AddCommand(initCmd)
}
