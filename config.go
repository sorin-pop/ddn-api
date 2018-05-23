package main

import (
	"github.com/djavorszky/ddn-common/logger"
)

// Config to hold the database server and ddn server configuration
type Config struct {
	DBAddress         string   `toml:"db-addr" required:"true"`
	DBUser            string   `toml:"db-username" required:"true"`
	DBPass            string   `toml:"db-userpass"`
	DBName            string   `toml:"db-name" required:"true"`
	ServerHost        string   `toml:"server-host" required:"true"`
	SMTPAddr          string   `toml:"smtp-host"`
	SMTPUser          string   `toml:"smtp-user"`
	SMTPPass          string   `toml:"smtp-password"`
	EmailSender       string   `toml:"email-sender"`
	AdminEmail        []string `toml:"admin-emails"`
	MountLoc          string   `toml:"mount-loc"`
	WebPushEnabled    bool     `toml:"webpush-enabled"`
	WebPushSubscriber string   `toml:"webpush-subscriber"`
	VAPIDPrivateKey   string   `toml:"vapid-private-key"`
	GoogleAnalyticsID string   `toml:"google-analytics-id"`
	LogLevel          string   `toml:"log-level"`
	StartupDelay      string   `toml:"startup-delay"`
}

// Print prints the configuration to the log.
func (c Config) Print() {
	logger.Info("Database Address:\t\t%s", c.DBAddress)
	logger.Info("Database User:\t\t%s", c.DBUser)
	logger.Info("Database Name:\t\t%s", c.DBName)

	logger.Info("Server Host:\t\t%s", c.ServerHost)

	if c.SMTPAddr != "" && c.EmailSender != "" {
		logger.Info("Admin email:\t\t%s", c.AdminEmail)
		logger.Info("Server configured to send emails.")
	}

	if c.GoogleAnalyticsID != "" {
		logger.Info("Google analytics enabled.")
	}
}
