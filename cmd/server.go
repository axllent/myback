package cmd

import (
	"runtime"

	"github.com/axllent/myback/logger"
	"github.com/axllent/myback/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the MyBack server",
	Long: `Run the MyBack server.

The MyBack server starts a HTTP(S) API service which the MyBack client connects to, and
uses MySQL's user database to authenticate. Internally, the server uses mysqldump to
stream the MySQL dumps to the client.

If you assign both '--ssl-key' & '--ssl-cert' then your server will listen with HTTPS,
otherwise HTTP is used.

Note: MyBack is not responsible for renewing certificates such as Lets Encrypt. Certficates
should be renewed using other methods, after which the MyBack server should be restarted.

Documentation, issues & support:
  https://github.com/axllent/myback`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {

		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			logger.Level = "vvv"
		}

		return server.Listen(Version)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	mysqldump := "mysqldump"

	if runtime.GOOS == "windows" {
		mysqldump = "mysqldump.exe"
	}

	serverCmd.Flags().
		StringVar(&server.Config.MySQLHost, "mysql-host", "localhost", "MySQL server host")
	serverCmd.Flags().
		IntVar(&server.Config.MySQLPort, "mysql-port", 3306, "MySQL server port")
	serverCmd.Flags().
		StringVar(&server.Config.MySQLDump, "mysqldump", mysqldump, "mysqldump command")
	serverCmd.Flags().
		StringVar(&server.Config.Listen, "listen", "0.0.0.0:3307", "listen on interface:port")
	serverCmd.Flags().
		StringVar(&server.Config.SSLCert, "ssl-cert", "", "SSL certificate (optional, must be used with --ssl-key)")
	serverCmd.Flags().
		StringVar(&server.Config.SSLKey, "ssl-key", "", "SSL private key (optional, must be used with --ssl-cert)")
	serverCmd.Flags().
		StringVar(&server.Config.LimitUsers, "users", "", "limit to users (comma-separated)")
	serverCmd.Flags().
		StringVar(&server.Config.LimitIPs, "ips", "", "limit to ips (comma-separated)")
	serverCmd.Flags().
		BoolVarP(&logger.ShowTimestamps, "show-timestamps", "t", false, "show timestamps in log output")
	serverCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
