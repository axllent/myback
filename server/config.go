package server

import (
	"os"
	"strings"

	"github.com/axllent/myback/logger"
)

// Config for server
var Config struct {
	MySQLHost  string
	MySQLPort  int
	Listen     string
	MySQLDump  string
	SSLCert    string
	SSLKey     string
	LimitIPs   string // comma-separated
	LimitUsers string // comma-separated
}

func checkConfig() {
	if Config.SSLCert == "" && Config.SSLKey != "" ||
		Config.SSLCert != "" && Config.SSLKey == "" {
		logger.Log().Error("You must speficy both an SSL certificate & SSL private key")
		os.Exit(1)
	}

	if Config.SSLCert != "" {
		if _, err := os.Stat(Config.SSLCert); err != nil {
			logger.Log().Errorf("SSL certificate %s does not exist", Config.SSLCert)
			os.Exit(1)
		}
		if _, err := os.Stat(Config.SSLKey); err != nil {
			logger.Log().Errorf("SSL private key %s does not exist", Config.SSLKey)
			os.Exit(1)
		}
	}

	if _, err := which(Config.MySQLDump); err != nil {
		logger.Log().Errorf("%s not found in your path", Config.MySQLDump)
		os.Exit(1)
	}

	if Config.LimitIPs != "" {
		ips := strings.Split(Config.LimitIPs, ",")
		for _, ip := range ips {
			limitToIPs[strings.TrimSpace(ip)] = true
		}

		logger.Log().Infof("Restricting to ip addresses: %s", strings.Join(ips, ", "))
	}

	if Config.LimitUsers != "" {
		users := strings.Split(Config.LimitUsers, ",")
		for _, user := range users {
			limitToUsers[strings.TrimSpace(user)] = true
		}

		logger.Log().Infof("Restricting to users: %s", strings.Join(users, ", "))
	}
}
