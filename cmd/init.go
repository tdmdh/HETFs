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

AUTHENTICATION SETUP
─────────────────────────────────────────────────────────────────
The IBKR Client Portal Gateway uses browser cookies for authentication.
To allow halalctl to use your authenticated session:

  1. Start the Gateway:   bin\run.bat root\conf.yaml
  2. Login in browser:    https://127.0.0.1:5000
  3. Open DevTools (F12) → Application → Cookies → https://127.0.0.1
  4. Copy the value of the 'x-sess-uuid' cookie
  5. Paste it into your .env file:
        SESSION_COOKIE=x-sess-uuid=<paste-value-here>
  6. Re-run:   go run . init
─────────────────────────────────────────────────────────────────`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cookie := os.Getenv("SESSION_COOKIE")
		if cookie == "" {
			fmt.Println("⚠️  SESSION_COOKIE is not set in your .env file.")
			fmt.Println("   After logging in at https://127.0.0.1:5000, copy the 'x-sess-uuid' cookie")
			fmt.Println("   from browser DevTools (F12 → Application → Cookies) and add it to .env:")
			fmt.Println("   SESSION_COOKIE=x-sess-uuid=<paste-value-here>")
			fmt.Println()
		}

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
		var promptShown bool

		for i := 1; i <= maxRetries; i++ {
			status, err := ibkrClient.CheckAuthStatus(cmd.Context())

			if err == nil && status.Authenticated {
				fmt.Println("\n✅ Gateway is AUTHENTICATED. Ready to serve requests.")
				authSuccess = true
				break
			}

			if err == nil && !promptShown {
				fmt.Println("\n✅ Gateway is reachable and responding!")
				fmt.Println("⚠️  Gateway is NOT authenticated yet.")
				if cookie == "" {
					fmt.Println("   👉 Add SESSION_COOKIE to your .env file (see instructions above), then re-run init.")
					break // No point polling if no cookie is set
				}
				fmt.Printf("   Please visit: %s to log in to Interactive Brokers.\n   Waiting for you to log in", gatewayURL)
				promptShown = true
			}

			fmt.Print(".")
			if err != nil {
				authMessage = err.Error()
			} else {
				authMessage = "Not Authenticated"
			}
			time.Sleep(delay)
		}

		if !authSuccess {
			fmt.Printf("\n\n❌ Could not verify authentication within the timeout period.\n")
			if authMessage != "" {
				fmt.Printf("Last error: %s\n", authMessage)
			}
			if cookie == "" {
				fmt.Println("\nSolution: Add SESSION_COOKIE to .env (see the instructions in 'go run . init --help')")
			} else {
				fmt.Println("\nThe cookie may have expired. Please log in again at https://127.0.0.1:5000 and update SESSION_COOKIE in .env")
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
