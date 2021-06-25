package cmd

import (
	"github.com/axllent/myback/client"
	"github.com/axllent/myback/logger"
	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract <dir> [dir] [dir]",
	Short: "Generate a SQL file from backups directories",
	Long: `Generate a SQL file from backup directories.

This will take a directory (or directory) and generate a single SQL file.
It will recursively scan directories for underlying directories to find 
backed up SQL dumps.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")

		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			logger.Level = "vvv"
		}

		if err := client.ExtractPaths(output, args); err != nil {
			logger.Log().Error(err.Error())
		}

		logger.Log().Noticef("Wrote to %s", output)
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)

	extractCmd.Flags().StringP("output", "o", "", "output to file")
	if err := extractCmd.MarkFlagRequired("output"); err != nil {
		logger.Log().Error(err.Error())
	}

	extractCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
