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
	Short: "Verifies connectivity and authentication to a running IBKR Gateway",
	Long: `Verifies the /auth/status endpoint of a running IBKR Client Portal Gateway.
If the gateway is active but unauthenticated, it prompts the user to log in.
If the gateway is not running, it attempts to launch it via docker-compose.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("⏳ Polling Gateway at %s...\n", gatewayURL)

		// Quick initial check to see if it's already running
		_, err := ibkrClient.CheckAuthStatus(cmd.Context())
		if err != nil {
			fmt.Println("\n⚠️ Gateway is not responding. Attempting to start it via docker-compose...")
			
			upCmd := exec.Command("docker-compose", "up", "-d")
			upCmd.Stdout = os.Stdout
			upCmd.Stderr = os.Stderr
			if err := upCmd.Run(); err != nil {
				fmt.Println("❌ Failed to start docker-compose. Please ensure Docker is running and docker-compose.yml is configured.")
				return fmt.Errorf("docker-compose up failed: %w", err)
			}
			
			fmt.Println("🚀 Docker container starting! Waiting up to 60 seconds for gateway to boot...")
		} else {
			fmt.Println("✅ Gateway is already accessible. Checking authentication...")
		}

		maxRetries := 30
		delay := 2 * time.Second
		
		var authSuccess bool
		var authMessage string
		
		for i := 1; i <= maxRetries; i++ {
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
			
			fmt.Print(".")
			authMessage = err.Error()
			time.Sleep(delay)
		}

		if !authSuccess {
			fmt.Printf("\n\n❌ Could not verify authentication within the timeout period.\n")
			fmt.Printf("Last connection error: %s\n", authMessage)
			fmt.Printf("\nPlease ensure you have correctly configured TWS_USERID and TWS_PASSWORD in docker-compose.yml.\n")
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
