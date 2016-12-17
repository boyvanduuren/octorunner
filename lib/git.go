package git

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	//"github.com/google/go-github/github"
	"net/http"
)

const (
	EVENTHEADER     = "X-GitHub-Event"
	FORWARDEDHEADER = "X-Forwarded-For"
)

type hookPayload struct {
	ref, before, after, compare string
	created, deleted, forced    bool
	Repository                  struct {
		id              int
		name, full_name string
		private         bool
	} `json:"repository"`
	Pusher struct {
		name, email string
	} `json:"pusher"`
	Sender struct {
		login string
		id    int
	} `json:"sender"`
}

// HandleWebhook is called when we receive a request on our listener and is responsible
// for decoding the payload and passing it to the appropriate handler for that particular event.
// If the received event is not supported we log an error and return without doing anything.
func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Map Github webhook events to functions that handle them
	supportedEvents := map[string]func(hookPayload){
		"push": handlePush,
	}

	log.Info("Received request on listener")
	// Request might be proxied, so check if there's an X-Forwarded-For header
	forwardedFor := r.Header.Get(FORWARDEDHEADER)
	var remoteAddr string
	if forwardedFor != "" {
		remoteAddr = forwardedFor
	} else {
		remoteAddr = r.RemoteAddr
	}
	log.Debug("Request from " + r.UserAgent() + " at " + remoteAddr)

	// Check which event we received and assign the appropriate handler to eventHandler
	var eventHandler func(hookPayload)
	event := r.Header.Get(EVENTHEADER)
	if event == "" {
		log.Error("Header \"" + EVENTHEADER + "\" not set, returning")
		return
	} else if val, exists := supportedEvents[event]; exists {
		eventHandler = val
		log.Debug("Found appropriate handler for " + event + "\" event")
	} else {
		log.Error("Received \"" + EVENTHEADER + "\", but found no supporting handler for \"" +
			event + "\" event, returning")
		return
	}

	// Try to decode the payload
	jsonDecoder := json.NewDecoder(r.Body)
	var payload hookPayload
	err := jsonDecoder.Decode(&payload)
	if err != nil {
		log.Error("Error while decoding payload: ", err)
		return
	}

	log.Debug("Decoded payload to ", payload)
	eventHandler(payload)
}

// Handle a push event to a Github repository. We will need to look at the settings for octorunner
// in this repository and take action accordingly.
func handlePush(payload hookPayload) {
	// do something
}
