package cmd

import (
	"fmt"
	"os"

	"github.com/axllent/myback/client"
	"github.com/axllent/myback/logger"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup <client-config>",
	Short: "Backup from a server running the MyBack server",
	Long: `Backups up from a server running the MyBack server.

Documentation, issues & support:
  https://github.com/axllent/myback`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if err := client.ParseConfig(args[0]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if client.Config.Repo == "" {
			fmt.Println("Config repo not set")
			os.Exit(1)
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			logger.Level = "vvvv"
		}

		if err := client.CreateDirIfNotExists(client.Config.Repo); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if errors := client.Backup(); len(errors) > 0 {
			for _, err := range errors {
				fmt.Println(err)
			}
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().
		BoolVarP(&logger.ShowTimestamps, "show-timestamps", "t", false, "show timestamps in output")
	backupCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
