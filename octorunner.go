package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/boyvanduuren/octorunner/lib/auth"
	"github.com/boyvanduuren/octorunner/lib/git"
	"github.com/boyvanduuren/octorunner/lib/persist"
	"github.com/boyvanduuren/octorunner/lib/webapi/app"
	"github.com/boyvanduuren/octorunner/lib/webapi/controllers"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/logrus"
	"github.com/goadesign/goa/middleware"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
)

const (
	configFile          = "config"
	configPath          = "."
	logLevel            = "loglevel"
	logLevelDefault     = "info"
	webServer           = "web.server"
	webServerDefault    = "127.0.0.1"
	webPort             = "web.port"
	webPortDefault      = "8080"
	webPath             = "web.path"
	webPathDefault      = "payload"
	databasePath        = "database.path"
	databasePathDefault = "octorunner.db"
)

// Main entry point for our program. Used to read and set the configuration we'll be using, and setup a webserver.
func main() {
	LOGMAP := map[string]log.Level{
		"debug": log.DebugLevel,
		"error": log.ErrorLevel,
		"fatal": log.FatalLevel,
		"info":  log.InfoLevel,
	}

	log.Info("Starting octorunner")

	// Configure viper to read config from the environment
	// We'll use a EnvKeyReplacer so OCTORUNNER_GIT_APIKEY
	// overrides git.apikey defined in a config file
	viper.SetEnvPrefix(git.EnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// A config file might exist in the same dir as the octorunner binary, but is not required
	viper.SetConfigName(configFile)
	viper.AddConfigPath(configPath)
	viper.ReadInConfig()

	// Set some defaults
	viper.SetDefault(logLevel, logLevelDefault)
	viper.SetDefault(webServer, webServerDefault)
	viper.SetDefault(webPort, webPortDefault)
	viper.SetDefault(webPath, webPathDefault)
	viper.SetDefault(databasePath, databasePathDefault)

	// Set log level
	logLevel := strings.ToLower(viper.GetString(logLevel))
	if val, exists := LOGMAP[logLevel]; exists {
		log.Info("Setting log level to " + logLevel)
		log.SetLevel(val)
	} else {
		log.Error("Loglevel was set to invalid value " + logLevel + ", defaulting to " + logLevelDefault)
	}

	// Get the webserver configuration
	webServer := viper.GetString(webServer)
	webPort := viper.GetString(webPort)
	webPath := viper.GetString(webPath)

	// See if the database exists
	database := viper.GetString(databasePath)
	// Setup connection pool
	err := persist.OpenDatabase(database, &persist.DBConn)
	if err != nil {
		log.Panicf("Cannot setup connection to database: %q", err)
	}

	// Get authentication data for git
	var repositories map[string]auth.Repository
	err = viper.UnmarshalKey("repositories", &repositories)
	if err != nil {
		log.Info("No auth data for repositories found")
	} else {
		repos := make([]string, len(repositories))
		for k := range repositories {
			repos = append(repos, k)
		}
		log.Info("Auth data found for the following repos: ", repos, " (this excludes ENV vars)")
		git.Auth = auth.SimpleAuth{Store: repositories}
	}

	// Capture os.Interrupt so we can close the db connection
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		_ = <-sigChan
		log.Info("Received SIGINT, closing database connection")
		if persist.DBConn.Connection != nil {
			persist.DBConn.Connection.Close()
			os.Exit(0)
		}
	}()

	// Setup our webapi

	// Create service
	service := goa.New("octorunner")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.WithLogger(goalogrus.New(log.New()))
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount our Github payload handler
	service.Mux.Handle("POST", "/"+webPath, git.HandleWebhook)

	// Mount "job" controller
	c := controllers.NewJobController(service)
	app.MountJobController(service, c)
	// Mount "project" controller
	c2 := controllers.NewProjectController(service)
	app.MountProjectController(service, c2)

	// Start service
	if err := service.ListenAndServe(webServer + ":" + webPort); err != nil {
		service.LogError("startup", "err", err)
	}
}
