package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/boyvanduuren/octorunner/lib/auth"
	"github.com/boyvanduuren/octorunner/lib/git"
	"github.com/boyvanduuren/octorunner/lib/persist"
	"github.com/spf13/viper"
	"net/http"
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
	err := persist.OpenDatabase(database, &persist.Connection)
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

	// Start listening for webhooks
	listener := webServer + ":" + webPort
	log.Info("Opening listener on " + listener + ", handling requests at /" + webPath)
	http.HandleFunc("/"+webPath, git.HandleWebhook)
	err = http.ListenAndServe(listener, nil)
	if err != nil {
		log.Fatal("Error while opening listener: ", err)
	}
}
