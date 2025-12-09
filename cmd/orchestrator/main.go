package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "orchestrator",
	Short: "Cloud DR Orchestrator - Backup and restore tool for Oracle Cloud",
	Long: `A disaster recovery orchestrator that manages backups of PostgreSQL databases
and files, storing them securely in Oracle Cloud Object Storage.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
