package main

import (
	"./lib"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"strings"
)

const (
	ENVPREFIX          = "githubrunner"
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

	log.Info("Starting githubrunner")

	// Configure viper to read config from the environment
	// We'll use a EnvKeyReplacer so GITHUBRUNNER_GIT_APIKEY
	// overrides git.apikey defined in a config file
	viper.SetEnvPrefix(ENVPREFIX)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// A config file might exist in the same dir as the githubrunner binary, but is not required
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

	listener := webServer + ":" + webPort
	log.Info("Opening listener on " + listener + ", handling requests at /" + webPath)
	http.HandleFunc("/"+webPath, git.HandleWebhook)
	err := http.ListenAndServe(listener, nil)
	if err != nil {
		log.Fatal("Error while opening listener: ", err)
	}
}
