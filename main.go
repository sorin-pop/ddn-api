package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/djavorszky/ddn-api/database"
	"github.com/djavorszky/ddn-api/database/mysql"
	"github.com/djavorszky/ddn-api/mail"
	"github.com/djavorszky/ddn-common/brwsr"
	"github.com/djavorszky/ddn-common/logger"
	"github.com/djavorszky/sutils"
	"github.com/kelseyhightower/envconfig"
)

var (
	workdir   string
	config    Config
	db        database.BackendConnection
	version   string
	buildTime string
	commit    string
)

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

	logLevel, err := logger.Parse(config.LogLevel)
	if err != nil {
		logLevel = logger.INFO
	}

	logger.Level = logLevel

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

	db = &mysql.DB{
		Address:  config.DBAddress,
		User:     config.DBUser,
		Pass:     config.DBPass,
		Database: config.DBName,
	}

	if config.StartupDelay != "" {
		d, err := time.ParseDuration(config.StartupDelay)
		if err != nil {
			logger.Fatal("Invalid startup delay: %v", config.StartupDelay)
		}

		logger.Info("Delaying startup for %s", d)

		time.Sleep(d)
	}

	err = db.ConnectAndPrepare()
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Database connection established")

	if config.SMTPAddr != "" {
		if config.SMTPUser != "" {
			err = mail.Init(config.SMTPAddr, config.SMTPUser, config.SMTPPass, config.EmailSender)
		} else {
			err = mail.InitNoAuth(config.SMTPAddr, config.EmailSender)
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

	port := strings.Split(config.ServerHost, ":")[1]

	logger.Info("Starting to listen on port %s", port)

	port = fmt.Sprintf(":%s", port)

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

func loadPropertiesFromEnv() error {
	err := envconfig.Process("ddn", &config)
	if err != nil {
		return fmt.Errorf("reading from env: %v", err)
	}

	return nil
}
