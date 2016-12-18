package git

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	authentication "github.com/boyvanduuren/octorunner/lib/auth"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"net/url"
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
		Owner    struct {
			Name string `json:"name"`
		} `json:"owner"`
		Private bool
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
		log.Debug("Found appropriate handler for \"" + event + "\" event")
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

	// The repository that this payload is for might have a secret configured, in which case we expect
	// a signature with the payload. The given signature then needs to match a signature we calculate ourselves.
	// Only then will we call our handler, else we'll log an error and return
	repoSecret := Auth.RequestSecret(payload.Repository.FullName)
	if len(repoSecret) == 0 {
		log.Error("No secret was configured, cannot verify their signature")
	} else {
		signature := r.Header.Get(SIGNATUREHEADER)
		if signature == "" {
			log.Error("Expected signature for payload, but none given")
			return
		}
		log.Debug("Received signature " + signature)
		calculatedSignature := "sha1=" + authentication.CalculateSignature(repoSecret, payloadBody)
		log.Debug("Calculated signature ", calculatedSignature)
		if !authentication.CompareSignatures([]byte(signature), []byte(calculatedSignature)) {
			log.Error("Signatures didn't match")
			return
		}
	}
	eventHandler(payload)
}

// Handle a push event to a Github repository. We will need to look at the settings for octorunner
// in this repository and take action accordingly.
func handlePush(payload hookPayload) {
	log.Info("Handling received push event")

	repoPrivate := payload.Repository.Private
	repoFullName := payload.Repository.FullName
	repoToken := Auth.RequestToken(repoFullName)

	log.Info("Repository \"" + repoFullName + "\" was pushed to")

	// In case of a private repository we'll need to see if we have credentials for it, because if we don't
	// we cannot download the repository from github
	if repoPrivate {
		log.Debug("Repository is private, looking up credentials")
		if repoToken == nil {
			log.Error("No token found for repository \"" + repoFullName + "\", returning")
			return
		}
	}

	repoName := payload.Repository.Name
	repoOwner := payload.Repository.Owner.Name
	commitId := payload.After
	getArchive(repoName, repoOwner, commitId, repoToken)
}

func getArchive(repoName string, repoOwner string, commitId string, repoToken *oauth2.Token) string {
	const GITHUB_ARCHIVE_URL = "https://github.com/%s/%s/archive/%s.zip"
	const GITHUB_ARCHIVE_FORMAT = "zipball"
	var archiveUrl *url.URL
	var err error

	log.Info("Downloading archive of latest commit in push")
	if repoToken == nil {
		// no repoToken, so this is a public repository
		archiveUrl, err = url.Parse(fmt.Sprintf(GITHUB_ARCHIVE_URL, repoOwner, repoName, commitId))
		if err != nil {
			log.Error("Error while constructing archive URL: ", err)
			return ""
		}
	} else {
		gitClient := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(repoToken)))
		log.Debug("Getting archive URL for \"" + repoOwner + "/" + repoName + "\", ref \"" + commitId + "\"")
		archiveUrl, _, err = gitClient.Repositories.GetArchiveLink(repoOwner, repoName, GITHUB_ARCHIVE_FORMAT,
			&github.RepositoryContentGetOptions{Ref: commitId})
		if err != nil {
			log.Error("Error while getting archive URL: ", err)
			return ""
		}
	}

	log.Debug("Found archive URL ", archiveUrl)

	return "stub"
}
