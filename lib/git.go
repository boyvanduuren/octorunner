package git

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

const (
	EVENTHEADER     = "X-GitHub-Event"
	FORWARDEDHEADER = "X-Forwarded-For"
)

type hookPayload struct {
	Username, Password string
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

// Handle a push event to a Github repository. We will need to look at the settings for githubrunner
// in this repository and take action accordingly.
func handlePush(payload hookPayload) {
	// do something
}
