package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/ddn-api/database"
	"github.com/djavorszky/ddn-api/database/mysql"
	"github.com/djavorszky/ddn-api/database/sqlite"
	"github.com/djavorszky/ddn-api/mail"
	"github.com/djavorszky/ddn-common/brwsr"
	"github.com/djavorszky/ddn-common/logger"
	"github.com/djavorszky/sutils"
)

var (
	workdir string
	config  Config
	db      database.BackendConnection
)

var version = ""

func main() {
	path, _ := filepath.Abs(os.Args[0])
	workdir = filepath.Dir(path)

	defer func() {
		if p := recover(); p != nil {
			if len(config.AdminEmail) != 0 {
				for _, addr := range config.AdminEmail {
					mail.Send(addr, "[FATAL] Cloud DB server panicked", fmt.Sprintf("%v", p))
				}
			}
		}
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-c
		// Received kill
		logger.Fatal("Received signal to terminate.")
	}()

	var err error
	confLocation := flag.String("p", "env", "Specify whether to read a configuration from a file (e.g. server.conf) or from environment variables.")
	logname := flag.String("l", "std", "Specify the log's filename. By default, logs to the terminal.")

	flag.Parse()

	if *logname != "std" {
		if _, err = os.Stat(*logname); err == nil {
			rotated := fmt.Sprintf("%s.%s", *logname, time.Now().Format("2006-01-02_03:04"))

			logger.Debug("Rotated logfile to %s", rotated)

			os.Rename(*logname, rotated)
		}

		logOut, err := os.OpenFile(*logname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("error opening file %s, will continue logging to stderr: %s", *logname, err.Error())
			logOut = os.Stderr
		}
		defer logOut.Close()

		log.SetOutput(logOut)
	}

	err = loadProperties(*confLocation)
	if err != nil {
		logger.Fatal("Failed loading configuration: %v", err)
	}

	logger.Info("Version: %s", version)

	logger.Info("Starting with properties:")

	config.Print()

	if config.MountLoc != "" {
		err = brwsr.Mount(config.MountLoc)
		if err != nil {
			logger.Warn("Couldn't mount folder: %s", err)

			config.MountLoc = ""
		} else {
			logger.Info("Mounted folder %q", config.MountLoc)
		}
	}

	switch config.DBProvider {
	case "mysql":
		db = &mysql.DB{
			Address:  config.DBAddress,
			Port:     config.DBPort,
			User:     config.DBUser,
			Pass:     config.DBPass,
			Database: config.DBName,
		}
	case "sqlite":
		db = &sqlite.DB{DBLocation: config.DBAddress}
	default:
		logger.Fatal("Unknown database provider: %s", config.DBProvider)
	}

	err = db.ConnectAndPrepare()
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Database connection established")

	if config.SMTPAddr != "" {
		if config.SMTPUser != "" {
			err = mail.Init(config.SMTPAddr, config.SMTPPort, config.SMTPUser, config.SMTPPass, config.EmailSender)
		} else {
			err = mail.InitNoAuth(config.SMTPAddr, config.SMTPPort, config.EmailSender)
		}

		if err != nil {
			logger.Warn("Mail failed to initialize: %v", err)
		} else {
			logger.Info("Mail initialized")
		}
	}

	if config.WebPushEnabled {
		if !sutils.Present(config.VAPIDPrivateKey, config.WebPushSubscriber) {
			logger.Error("WebPush is enabled but no private key / subscriber email specified! Disabling WebPush.")
			config.WebPushEnabled = false
		} else {
			logger.Info("WebPush is enabled.")
		}
	}

	// Start maintenance goroutine
	go maintain()

	// Start agent checker goroutine
	go checkAgents()

	logger.Info("Starting to listen on port %s", config.ServerPort)

	port := fmt.Sprintf(":%s", config.ServerPort)

	logger.Error("%v", http.ListenAndServe(port, Router()))

	if len(config.AdminEmail) != 0 {
		for _, addr := range config.AdminEmail {
			mail.Send(addr, "[Cloud DB] Server went down", fmt.Sprintf(`<p>Cloud DB down for some reason.</p>`))
		}
	}
}

func loadProperties(confLocation string) error {
	if confLocation != "env" {
		return loadPropertiesFromFile(confLocation)
	}

	return loadPropertiesFromEnv()
}

func loadPropertiesFromFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file doesn't exist: %s", filename)
	}

	if _, err := toml.DecodeFile(filename, &config); err != nil {
		return fmt.Errorf("couldn't read configuration file: %v", err)
	}

	return nil
}

const (
	envDBProvider        = "DB_PROVIDER"
	envDBAddress         = "DB_ADDRESS"
	envDBName            = "DB_NAME"
	envDBUser            = "DB_USER"
	envDBPassword        = "DB_PASSWORD"
	envServerHost        = "SERVER_HOST"
	envSMTPAddress       = "SMTP_ADDRESS"
	envSMTPUser          = "SMTP_USER"
	envSMTPPassword      = "SMTP_PASSWORD"
	envAdminEmails       = "ADMIN_EMAILS"
	envMountLocation     = "MOUNT_LOCATION"
	envWebpushEnabled    = "WEBPUSH_ENABLED"
	envWebpushSubscriber = "WEBPUSH_SUBSCRIBER"
	envWebpushPublicKey  = "WEBPUSH_PUBLIC_KEY"
	envWebpushPrivateKey = "WEBPUSH_PRIVATE_KEY"
	envGoogleAnalyticsID = "GOOGLE_ANALYTICS_ID"
)

func loadPropertiesFromEnv() error {
	// Provider
	dbProvider, err := loadRequiredProperty(envDBProvider)
	if err != nil {
		return err
	}
	config.DBProvider = dbProvider

	// DB Address
	dbAddress, err := loadRequiredProperty(envDBAddress)
	if err != nil {
		return err
	}

	addr := strings.Split(dbAddress, ":")
	if len(addr) == 1 {
		return fmt.Errorf("%q is missing the port", envDBAddress)
	}
	config.DBAddress = addr[0]
	config.DBPort = addr[1]

	// DB User and Password
	user, err := loadRequiredProperty(envDBUser)
	if err != nil {
		return err
	}
	config.DBUser = user

	dbname, err := loadRequiredProperty(envDBName)
	if err != nil {
		return err
	}
	config.DBName = dbname

	config.DBPass = loadOptionalProperty(envDBPassword)

	serverHost, err := loadRequiredProperty(envServerHost)
	if err != nil {
		return err
	}

	host := strings.Split(serverHost, ":")
	if len(host) == 1 {
		return fmt.Errorf("%q is missing the port", serverHost)
	}
	config.ServerHost = host[0]
	config.ServerPort = host[1]

	smtpAddr := loadOptionalProperty(envSMTPAddress)
	if smtpAddr != "" {

		smtpAddrArr := strings.Split(smtpAddr, ":")
		if len(smtpAddrArr) == 1 {
			return fmt.Errorf("%q is missing the port", smtpAddrArr)
		}

		config.SMTPAddr = smtpAddrArr[0]
		portNum, err := strconv.Atoi(smtpAddrArr[1])
		if err != nil {
			return fmt.Errorf("failed converting smtp port to int: %v", err)
		}

		config.SMTPPort = portNum
	}

	config.SMTPUser = loadOptionalProperty(envSMTPUser)
	config.SMTPPass = loadOptionalProperty(envSMTPPassword)

	adminEmails := loadOptionalProperty(envAdminEmails)

	config.AdminEmail = strings.Split(adminEmails, ",")

	config.MountLoc = loadOptionalProperty(envMountLocation)

	config.WebPushEnabled = loadOptionalProperty(envWebpushEnabled) == "true"
	config.WebPushSubscriber = loadOptionalProperty(envWebpushSubscriber)
	config.VAPIDPrivateKey = loadOptionalProperty(envWebpushPrivateKey)

	config.GoogleAnalyticsID = loadOptionalProperty(envGoogleAnalyticsID)

	return nil
}

func loadRequiredProperty(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("required environment variable missing: %s", key)
	}

	return val, nil
}

func loadOptionalProperty(key string) string {
	return os.Getenv(key)
}
