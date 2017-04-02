package git

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	authentication "github.com/boyvanduuren/octorunner/lib/auth"
	"github.com/boyvanduuren/octorunner/lib/common"
	"github.com/boyvanduuren/octorunner/lib/pipeline"
	"github.com/docker/docker/client"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"github.com/boyvanduuren/octorunner/lib/persist"
)

const (
	eventHeader     = "X-GitHub-Event"
	forwardedHeader = "X-Forwarded-For"
	signatureHeader = "X-Hub-Signature"
	tmpDirPrefix    = "octorunner-"
	tmpFilePrefix   = "archive-"
	pipelineFile    = ".octorunner"
	EnvPrefix       = "octorunner"
	envRepoToken    = "%s_%s_TOKEN"
	envRepoSecret   = "%s_%s_SECRET"
)

var Auth authentication.Method

const repositoryData string = "repositoryData"

type hookPayload struct {
	Ref, Before, After, Compare string
	Created, Deleted, Forced    bool
	Repository                  struct {
		ID       int
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
		ID    int
	} `json:"sender"`
}

// HandleWebhook is called when we receive a request on our listener and is responsible
// for decoding the payload and passing it to the appropriate handler for that particular event.
// If the received event is not supported we log an error and return without doing anything.
func HandleWebhook(w http.ResponseWriter, r *http.Request, v url.Values) {
	// Map Github webhook events to functions that handle them
	supportedEvents := map[string]func(hookPayload){
		"push": handlePush,
	}

	// Return 200 to the client
	w.WriteHeader(200)

	log.Info("Received request on listener")
	// Request might be proxied, so check if there's an X-Forwarded-For header
	forwardedFor := r.Header.Get(forwardedHeader)
	var remoteAddr string
	if forwardedFor != "" {
		remoteAddr = forwardedFor
	} else {
		remoteAddr = r.RemoteAddr
	}
	log.Debug("Request from " + r.UserAgent() + " at " + remoteAddr)

	// Check which event we received and assign the appropriate handler to eventHandler
	var eventHandler func(hookPayload)
	event := r.Header.Get(eventHeader)
	if event == "" {
		log.Error("Header \"" + eventHeader + "\" not set, returning")
		return
	} else if val, exists := supportedEvents[event]; exists {
		eventHandler = val
		log.Debug("Found appropriate handler for \"" + event + "\" event")
	} else {
		log.Error("Received \"" + eventHeader + "\", but found no supporting handler for \"" +
			event + "\" event, returning")
		return
	}

	// Read the body of the request
	payloadBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Errorf("Error while reading payload: %v", err)
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
		// We need to manually search the env, because viper doesn't seem to load these
		// environment vars into the config.
		repoSecretEnvKey := fmt.Sprintf(envRepoSecret, strings.ToUpper(EnvPrefix), payload.Repository.FullName)
		log.Debugf("Couldn't find secret in config file, looking if environment var %q exists", repoSecretEnvKey)
		repoSecret = []byte(os.Getenv(repoSecretEnvKey))
		if len(repoSecret) == 0 {
			log.Error("No secret was configured, cannot verify their signature")
		} else {
			log.Debugf("Found secret for %q in environment", payload.Repository.FullName)
			signature := r.Header.Get(signatureHeader)
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
	}
	go eventHandler(payload)
}

// Handle a push event to a Github repository. We will need to look at the settings for octorunner
// in this repository and take action accordingly.
func handlePush(payload hookPayload) {
	log.Info("Handling received push event")

	// Create a context for this request
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repoFullName := payload.Repository.FullName
	repoToken := Auth.RequestToken(repoFullName)

	log.Info("Repository \"" + repoFullName + "\" was pushed to")

	// See if we have a token for this repository. If we don't we won't be able to set a status so we abort.
	// we cannot download the repository from github
	if repoToken == nil {
		// We need to manually search the env, because viper doesn't seem to load these
		// environment vars into the config.
		repoTokenEnvKey := fmt.Sprintf(envRepoToken, strings.ToUpper(EnvPrefix), payload.Repository.FullName)
		log.Debugf("Couldn't find token in config file, looking if environment var %q exists", repoTokenEnvKey)
		repoTokenEnvVal := os.Getenv(repoTokenEnvKey)
		if repoTokenEnvVal == "" {
			log.Errorf("Didn't find token for %q, this means we won't be able to set a status. Aborting.",
				repoFullName)
			return
		}
		log.Debugf("Found token for %q in environment", repoFullName)
		repoToken = &oauth2.Token{AccessToken: repoTokenEnvVal}
	}

	repoName := payload.Repository.Name
	repoOwner := payload.Repository.Owner.Name
	commitID := payload.After

	/*
	 When a commit is merged from a branch to another branch, the "after" ID is set to
	 "0000000000000000000000000000000000000000", and the "previous" ID is the ID of the commit being merged.
	 That commit will probably already have a state assigned, so we can just return
	*/
	if commitID == "0000000000000000000000000000000000000000" {
		log.Info("Not doing anything, this was a merge commit")
		return
	}

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(repoToken))
	gitClient := github.NewClient(httpClient)

	repoDir, err := getRepository(ctx, httpClient, gitClient, repoName, repoOwner, commitID, repoToken)
	if err != nil {
		log.Errorf("Error while downloading copy of repository: %v", err)
		return
	}

	ctx = context.WithValue(ctx, repositoryData, map[string]string{
		"fullName":   repoFullName,
		"commitId":   commitID,
		"fsLocation": repoDir,
	})

	repoPipeline, err := readPipelineConfig(repoDir)
	if err != nil {
		log.Errorf("Error while reading pipeline configuration: %v", err)
		return
	}

	// set state of commit to pending
	log.Debug("Setting state to pending")
	gitSetState(ctx, gitClient, "pending", repoOwner, repoName, commitID)

	// create Docker client
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Errorf("Error while creating connection to Docker: %q", err)
		gitSetState(ctx, gitClient, "error", repoOwner, repoName, commitID)
		return
	}
	defer cli.Close()

	exitcode, err := repoPipeline.Execute(ctx, cli, &persist.DBConn)
	if err != nil {
		log.Errorf("Error while executing pipeline: %v", err)
		gitSetState(ctx, gitClient, "error", repoOwner, repoName, commitID)
		return
	}

	log.Debugf("Pipeline returned %d, setting state accordingly", exitcode)
	if exitcode == 0 {
		gitSetState(ctx, gitClient, "success", repoOwner, repoName, commitID)
	} else {
		gitSetState(ctx, gitClient, "failure", repoOwner, repoName, commitID)
	}
}

func getRepository(ctx context.Context, httpClient *http.Client, gitClient *github.Client, repoName string, repoOwner string,
	commitID string, repoToken *oauth2.Token) (string, error) {
	const githubArchiveURL = "https://github.com/%s/%s/archive/%s.zip"
	const githubArchiveFormat = "zipball"
	var archiveURL *url.URL
	var err error

	log.Info("Downloading archive of latest commit in push")
	if repoToken == nil {
		// no repoToken, so this is a public repository
		archiveURL, err = url.Parse(fmt.Sprintf(githubArchiveURL, repoOwner, repoName, commitID))
		if err != nil {
			return "", fmt.Errorf("Error while constructing archive URL: %v", err)
		}
	} else {
		log.Debug("Getting archive URL for \"" + repoOwner + "/" + repoName + "\", ref \"" + commitID + "\"")
		archiveURL, _, err = gitClient.Repositories.GetArchiveLink(ctx, repoOwner, repoName, githubArchiveFormat,
			&github.RepositoryContentGetOptions{Ref: commitID})
		if err != nil {
			return "", fmt.Errorf("Error while getting archive URL: %v", err)
		}
	}
	log.Debug("Found archive URL ", archiveURL)

	tmpDir, err := ioutil.TempDir("", tmpDirPrefix)
	if err != nil {
		log.Error()
		return "", fmt.Errorf("Error while creating temporary directory: %v", err)
	}

	log.Debug("Created temporary directory " + tmpDir)
	archivePath, err := downloadFile(httpClient, archiveURL, tmpDir)
	if err != nil {
		return "", fmt.Errorf("Error while downloading archive: %v", err)
	}
	log.Debug("Archive downloaded to ", archivePath.Name())

	repoDir, err := common.Unzip(archivePath.Name(), tmpDir)
	if err != nil {
		return "", fmt.Errorf("Error while unpacking archive: %v", err)
	}
	log.Debugf("Unpacked archive to %q", repoDir)

	// cleanup the archive
	archivePath.Close()
	os.Remove(archivePath.Name())

	// we should now have a copy of the repository at the latest commit
	if common.CheckDirNotExists(repoDir) {
		log.Error()
		return "", fmt.Errorf("Repository not found at expected directory \"%s\" after unpacking", repoDir)
	}
	log.Debug("Repository unpacked to ", repoDir)

	return repoDir, nil
}

func downloadFile(httpClient *http.Client, url *url.URL, downloadDirectory string) (*os.File, error) {
	log.Debug("Downloading \"" + url.String() + "\"")
	filePath, err := ioutil.TempFile(downloadDirectory, tmpFilePrefix)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Get(url.String())
	defer resp.Body.Close()
	n, err := io.Copy(filePath, resp.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Downloaded %d bytes", n)
	return filePath, nil
}

func readPipelineConfig(directory string) (pipeline.Pipeline, error) {
	var pipelineConfig pipeline.Pipeline
	pipelineConfigPath := path.Join(directory, strings.Join([]string{pipelineFile, ".yaml"}, ""))
	if _, err := os.Stat(pipelineConfigPath); os.IsNotExist(err) == true {
		pipelineConfigPath = path.Join(directory, strings.Join([]string{pipelineFile, ".yml"}, ""))
		if _, err := os.Stat(pipelineConfigPath); os.IsNotExist(err) == true {
			return pipelineConfig, errors.New("Couldn't find .octorunner.yaml or .octorunner.yml in repository")
		}
	}
	pipelineConfigBuf, err := ioutil.ReadFile(pipelineConfigPath)
	if err != nil {
		return pipelineConfig, fmt.Errorf("Error while reading from %s: %v", pipelineConfigPath, err)
	}

	pipelineConfig, err = pipeline.ParseConfig(pipelineConfigBuf)
	return pipelineConfig, err
}

func gitSetState(ctx context.Context, git *github.Client, state string, owner string, repo string, commit string) {
	repoStatusContext := "continuous-integration/octorunner"
	git.Repositories.CreateStatus(ctx, owner, repo, commit,
		&github.RepoStatus{
			State:   &state,
			Context: &repoStatusContext,
		},
	)
}
