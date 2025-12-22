package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Manage automated backup schedules using Cronify",
	Long: `Manage automated backup schedules using YAML configuration and Cronify.

Cronify is a YAML-driven tool for managing, validating, and deploying cron jobs.
Use this command to create, validate, and deploy backup schedules.

Example:
  orchestrator schedule init
  orchestrator schedule validate --file backup-schedule.yaml
  orchestrator schedule deploy --file backup-schedule.yaml`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate example backup schedule YAML file",
	Long: `Generate an example backup-schedule.yaml file with common backup patterns.

The generated file includes examples for:
- Daily backups (midnight)
- Weekly backups (Sunday morning)
- Monthly backups (1st of month)

Example:
  orchestrator schedule init
  orchestrator schedule init --output custom-schedule.yaml`,
	RunE: runInit,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate backup schedule YAML file",
	Long: `Validate a backup schedule YAML file using Cronify.

Checks:
- YAML syntax
- Cron expression validity
- Schedule simulation (next 5 runs)
- Environment variable availability
- Command file existence

Example:
  orchestrator schedule validate --file backup-schedule.yaml
  orchestrator schedule validate --file backup-schedule.yaml --simulate`,
	RunE: runValidate,
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy backup schedule to crontab",
	Long: `Deploy a backup schedule YAML file to crontab using Cronify.

This command:
1. Validates the YAML file
2. Converts it to crontab format
3. Deploys to system crontab

Example:
  orchestrator schedule deploy --file backup-schedule.yaml
  orchestrator schedule deploy --file backup-schedule.yaml --dry-run`,
	RunE: runDeploy,
}

var (
	scheduleFile     string
	scheduleOutput   string
	scheduleDryRun   bool
	scheduleSimulate bool
)

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.AddCommand(initCmd)
	scheduleCmd.AddCommand(validateCmd)
	scheduleCmd.AddCommand(deployCmd)

	// init flags
	initCmd.Flags().StringVarP(&scheduleOutput, "output", "o", "backup-schedule.yaml", "Output file path")

	// validate flags
	validateCmd.Flags().StringVarP(&scheduleFile, "file", "f", "", "Path to backup schedule YAML file (required)")
	validateCmd.Flags().BoolVar(&scheduleSimulate, "simulate", false, "Simulate next 5 runs")
	validateCmd.MarkFlagRequired("file")

	// deploy flags
	deployCmd.Flags().StringVarP(&scheduleFile, "file", "f", "", "Path to backup schedule YAML file (required)")
	deployCmd.Flags().BoolVar(&scheduleDryRun, "dry-run", false, "Preview crontab without deploying")
	deployCmd.MarkFlagRequired("file")
}

type ScheduleConfig struct {
	Jobs []Job `yaml:"jobs"`
}

type Job struct {
	Name     string            `yaml:"name"`
	Schedule string            `yaml:"schedule"`
	Command  string            `yaml:"command"`
	Env      map[string]string `yaml:"env,omitempty"`
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Printf("📝 Generating example backup schedule: %s\n", scheduleOutput)

	// Create example schedule
	schedule := ScheduleConfig{
		Jobs: []Job{
			{
				Name:     "daily-backup",
				Schedule: "0 0 * * *",
				Command:  "/usr/local/bin/orchestrator backup --name prod-db --db-name myapp --db-host localhost --db-user postgres --db-password secret --encrypt && /usr/local/bin/orchestrator upload --file /tmp/backup-*.tar.gz.encrypted --bucket my-bucket --compartment ocid1.compartment.oc1..xxx",
				Env: map[string]string{
					"BACKUP_ENCRYPTION_KEY": "your-encryption-key-here",
					"PATH":                  "/usr/local/bin:/usr/bin:/bin",
				},
			},
			{
				Name:     "weekly-backup",
				Schedule: "0 3 * * 0",
				Command:  "/usr/local/bin/orchestrator backup --name prod-db-weekly --db-name myapp --db-host localhost --db-user postgres --db-password secret --encrypt",
				Env: map[string]string{
					"BACKUP_ENCRYPTION_KEY": "your-encryption-key-here",
				},
			},
			{
				Name:     "monthly-backup",
				Schedule: "0 2 1 * *",
				Command:  "/usr/local/bin/orchestrator backup --name prod-db-monthly --db-name myapp --db-host localhost --db-user postgres --db-password secret --encrypt",
				Env: map[string]string{
					"BACKUP_ENCRYPTION_KEY": "your-encryption-key-here",
				},
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(&schedule)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Write to file
	err = os.WriteFile(scheduleOutput, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ Created %s\n\n", scheduleOutput)
	fmt.Println("📋 Example schedules:")
	fmt.Println("  • Daily backup:   Every day at midnight (0 0 * * *)")
	fmt.Println("  • Weekly backup:  Every Sunday at 3 AM (0 3 * * 0)")
	fmt.Println("  • Monthly backup: 1st of month at 2 AM (0 2 1 * *)")
	fmt.Println("\n⚠️  Edit the file to:")
	fmt.Println("  1. Replace database credentials")
	fmt.Println("  2. Set your encryption key")
	fmt.Println("  3. Update bucket and compartment IDs")
	fmt.Println("  4. Adjust schedules as needed")
	fmt.Println("\n Next steps:")
	fmt.Printf("  orchestrator schedule validate --file %s\n", scheduleOutput)
	fmt.Printf("  orchestrator schedule deploy --file %s\n", scheduleOutput)

	return nil
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Check if cronify is installed
	if !isCronifyInstalled() {
		return fmt.Errorf("cronify is not installed\n\nInstall Cronify:\n  git clone https://github.com/Kobeep/Cronify.git\n  cd Cronify\n  sudo ./install.sh\n\nOr install manually from: https://github.com/Kobeep/Cronify")
	}

	fmt.Printf("🔍 Validating schedule file: %s\n", scheduleFile)

	// Check if file exists
	if _, err := os.Stat(scheduleFile); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", scheduleFile)
	}

	// Build cronify command
	cronifyArgs := []string{"--file", scheduleFile}
	if scheduleSimulate {
		cronifyArgs = append(cronifyArgs, "--simulate")
	}

	// Run cronify validation
	cronifyCmd := exec.Command("cronify", cronifyArgs...)
	cronifyCmd.Stdout = os.Stdout
	cronifyCmd.Stderr = os.Stderr

	if err := cronifyCmd.Run(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Println("\n✓ Validation successful!")
	return nil
}

func runDeploy(cmd *cobra.Command, args []string) error {
	// Check if cronify is installed
	if !isCronifyInstalled() {
		return fmt.Errorf("cronify is not installed\n\nInstall Cronify:\n  git clone https://github.com/Kobeep/Cronify.git\n  cd Cronify\n  sudo ./install.sh\n\nOr install manually from: https://github.com/Kobeep/Cronify")
	}

	fmt.Printf("🚀 Deploying schedule file: %s\n", scheduleFile)

	// Check if file exists
	if _, err := os.Stat(scheduleFile); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", scheduleFile)
	}

	// Build cronify command
	cronifyArgs := []string{"--deploy", scheduleFile}
	if scheduleDryRun {
		cronifyArgs = append(cronifyArgs, "--simulate")
		fmt.Println("🔎 Dry-run mode: Preview only, no changes will be made\n")
	}

	// Run cronify deployment
	cronifyCmd := exec.Command("cronify", cronifyArgs...)
	cronifyCmd.Stdout = os.Stdout
	cronifyCmd.Stderr = os.Stderr

	if err := cronifyCmd.Run(); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	if !scheduleDryRun {
		fmt.Println("\n✓ Deployment successful! Backup schedules are now active.")
		fmt.Println("\n📋 View current crontab:")
		fmt.Println("  crontab -l")
	}

	return nil
}

func isCronifyInstalled() bool {
	_, err := exec.LookPath("cronify")
	return err == nil
}
