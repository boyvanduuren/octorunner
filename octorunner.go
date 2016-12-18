package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/boyvanduuren/octorunner/lib/auth"
	"github.com/boyvanduuren/octorunner/lib/git"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

const (
	ENVPREFIX          = "octorunner"
	CONFIGFILE         = "config"
	CONFIGPATH         = "."
	LOGLEVEL           = "loglevel"
	LOGLEVEL_DEFAULT   = "info"
	WEB_SERVER         = "web.server"
	WEB_SERVER_DEFAULT = "127.0.0.1"
	WEB_PORT           = "web.port"
	WEB_PORT_DEFAULT   = "8080"
	WEB_PATH           = "web.path"
	WEB_PATH_DEFAULT   = "payload"
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
	viper.SetEnvPrefix(ENVPREFIX)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// A config file might exist in the same dir as the octorunner binary, but is not required
	viper.SetConfigName(CONFIGFILE)
	viper.AddConfigPath(CONFIGPATH)
	viper.ReadInConfig()

	// Set some defaults
	viper.SetDefault(LOGLEVEL, LOGLEVEL_DEFAULT)
	viper.SetDefault(WEB_SERVER, WEB_SERVER_DEFAULT)
	viper.SetDefault(WEB_PORT, WEB_PORT_DEFAULT)
	viper.SetDefault(WEB_PATH, WEB_PATH_DEFAULT)

	// Set log level
	logLevel := strings.ToLower(viper.GetString(LOGLEVEL))
	if val, exists := LOGMAP[logLevel]; exists {
		log.Info("Setting log level to " + logLevel)
		log.SetLevel(val)
	} else {
		log.Error("Loglevel was set to invalid value " + logLevel + ", defaulting to " + LOGLEVEL_DEFAULT)
	}

	// Get the webserver configuration
	webServer := viper.GetString(WEB_SERVER)
	webPort := viper.GetString(WEB_PORT)
	webPath := viper.GetString(WEB_PATH)

	// Get authentication data for git
	var repositories map[string]auth.Repository
	err := viper.UnmarshalKey("repositories", &repositories)
	if err != nil {
		log.Info("No auth data for repositories found")
	} else {
		var repos []string = make([]string, len(repositories))
		for k := range repositories {
			repos = append(repos, k)
		}
		log.Info("Auth data found for the following repos: ", repos)
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
