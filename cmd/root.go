package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "myback",
	Short: "MySQL backup server and client",
	Long: `MyBack is a MySQL/MariaDB backup server and client.

The server (see 'myback server -h') will run on either the MySQL server itself, or a
server that can communicate with the MySQL/MariaDB server over TCP.

The client (see 'myback backup -h') will then dump the configured databases & tables
over HTTP(S).

Documentation, issues & support:
  https://github.com/axllent/myback`,
}

func Execute() {
	// hide the `help` command
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
