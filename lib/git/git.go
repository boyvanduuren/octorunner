package git

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	//	"github.com/google/go-github/github"
	"bytes"
	authentication "github.com/boyvanduuren/octorunner/lib/auth"
	"io/ioutil"
	"net/http"
)

const (
	EVENTHEADER     = "X-GitHub-Event"
	FORWARDEDHEADER = "X-Forwarded-For"
	SIGNATUREHEADER = "X-Hub-Signature"
)

var Auth authentication.AuthMethod

type hookPayload struct {
	Ref, Before, After, Compare string
	Created, Deleted, Forced    bool
	Repository                  struct {
		Id       int
		Name     string
		FullName string `json:"full_name"`
		Private  bool
	} `json:"repository"`
	Pusher struct {
		Name, Email string
	} `json:"pusher"`
	Sender struct {
		Login string
		Id    int
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

	// Read the body of the request
	payloadBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error("Error while reading payload: %v", err)
	} else {
		log.Debug("Received body ", string(payloadBody))
	}

	// Try to decode the payload
	jsonDecoder := json.NewDecoder(bytes.NewReader(payloadBody))
	var payload hookPayload
	err = jsonDecoder.Decode(&payload)
	if err != nil {
		log.Error("Error while decoding payload: ", err)
		return
	}
	log.Debug("Decoded payload to ", payload)

	// The payload might have an X-Hub-Signature header, which means the webhook has a secret, so we have to
	// validate the signature contained in the header
	signature := r.Header.Get(SIGNATUREHEADER)
	if signature != "" {
		log.Debug("Received signature \"" + signature + "\" for payload")
	}

	// If we actually received a signature, calculate our own and validate
	// On succesful validation call our handler, else return with an error log message
	if signature != "" {
		log.Debug("Received signature " + signature)
		repoSecret := Auth.RequestSecret(payload.Repository.FullName)
		if len(repoSecret) == 0 {
			log.Error("No secret was configured, cannot verify their signature")
			return
		}
		calculatedSignature := authentication.CalculateSignature(repoSecret, payloadBody)
		log.Debug("Calculated signature ", calculatedSignature)
		if authentication.CompareSignatures([]byte(signature), []byte("sha1="+calculatedSignature)) {
			eventHandler(payload)
		} else {
			log.Error("Signatures didn't match")
			return
		}
	}
}

// Handle a push event to a Github repository. We will need to look at the settings for octorunner
// in this repository and take action accordingly.
func handlePush(payload hookPayload) {
	log.Info("Handling received push event")

	repoPrivate := payload.Repository.Private
	repoFullName := payload.Repository.FullName
	var repoToken string

	log.Info("Repository \"" + repoFullName + "\" was pushed to")

	// In case of a private repository we'll need to see if we have credentials for it, because if we don't
	// we cannot download the repository from github
	if repoPrivate {
		log.Debug("Repository is private, looking up credentials")
		repoToken = Auth.RequestToken(repoFullName)
		if repoToken == "" {
			log.Error("No token found for repository \"" + repoFullName + "\", returning")
			return
		}
	}
}
