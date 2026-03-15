package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Verifies connectivity and authentication to a running IBKR Gateway",
	Long: `Verifies the /auth/status endpoint of a running IBKR Client Portal Gateway.
If the gateway is active but unauthenticated, it prompts the user to log in.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("⏳ Polling Gateway at %s (waiting up to 60 seconds)...\n", gatewayURL)

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
			fmt.Printf("\n\nCould not verify authentication within the timeout period.\n")
			fmt.Printf("Last connection error: %s\n", authMessage)
			fmt.Printf("\nPlease ensure you have started an IBKR Client Portal Gateway and navigate to %s to log in.\n", gatewayURL)
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
