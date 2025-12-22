package main

import (
	"fmt"

	"github.com/Kobeep/cloud-dr-orchestrator/pkg/encryption"
	"github.com/spf13/cobra"
)

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a new encryption key",
	Long: `Generate a secure random 256-bit encryption key for backup encryption.
The key is displayed as base64-encoded string that can be stored in environment
variables or configuration files.

Example:
  # Generate a new key
  orchestrator keygen

  # Save to environment variable
  export BACKUP_ENCRYPTION_KEY=$(orchestrator keygen)

  # Save to file
  orchestrator keygen > ~/.backup-key

Security notes:
  - Store the key securely (use environment variables or secret managers)
  - Never commit keys to version control
  - Backup your key! Lost keys = lost backups
  - Consider using different keys for different environments (dev/staging/prod)
`,
	RunE: runKeygen,
}

func init() {
	rootCmd.AddCommand(keygenCmd)
}

func runKeygen(cmd *cobra.Command, args []string) error {
	key, err := encryption.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	fmt.Println("🔑 Generated 256-bit encryption key:")
	fmt.Println(key)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT:")
	fmt.Println("   - Store this key securely!")
	fmt.Println("   - Never commit it to version control")
	fmt.Println("   - Backup the key (lost key = lost backups)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("   export BACKUP_ENCRYPTION_KEY=\"" + key + "\"")
	fmt.Println("   orchestrator backup --encrypt --encryption-key \"$BACKUP_ENCRYPTION_KEY\" ...")

	return nil
}
