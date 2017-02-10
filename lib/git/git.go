package git

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	authentication "github.com/boyvanduuren/octorunner/lib/auth"
	"github.com/boyvanduuren/octorunner/lib/pipeline"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"path/filepath"
	"archive/zip"
	"golang.org/x/net/context"
)

const (
	EVENTHEADER     = "X-GitHub-Event"
	FORWARDEDHEADER = "X-Forwarded-For"
	SIGNATUREHEADER = "X-Hub-Signature"
	TMPDIR_PREFIX   = "octorunner-"
	TMPFILE_PREFIX  = "archive-"
	PIPELINEFILE    = ".octorunner"
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
	supportedEvents := map[string]func(context.Context, hookPayload){
		"push": handlePush,
	}

	// Create a context for this request
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	var eventHandler func(context.Context, hookPayload)
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
	eventHandler(ctx, payload)
}

// Handle a push event to a Github repository. We will need to look at the settings for octorunner
// in this repository and take action accordingly.
func handlePush(ctx context.Context, payload hookPayload) {
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

	/*
	 When a commit is merged from a branch to another branch, the "after" ID is set to
	 "0000000000000000000000000000000000000000", and the "previous" ID is the ID of the commit being merged.
	 That commit will probably already have a state assigned, so we can just return
	*/
	if commitId == "0000000000000000000000000000000000000000" {
		return
	}

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(repoToken))
	gitClient := github.NewClient(httpClient)

	repoDir, err := getRepository(httpClient, gitClient, repoName, repoOwner, commitId, repoToken)
	if err != nil {
		log.Errorf("Error while downloading copy of repository: %v", err)
		return
	}

	ctx = context.WithValue(ctx, "repoData", map[string]string{
		"fullName":   repoFullName,
		"commitId":   commitId,
		"fsLocation": repoDir,
	})

	repoPipeline, err := readPipelineConfig(repoDir)
	if err != nil {
		log.Errorf("Error while reading pipeline configuration: %v", err)
		return
	}

	// set state of commit to pending
	log.Debug("Setting state to pending")
	gitSetState(gitClient, "pending", repoOwner, repoName, commitId)
	exitcode, err := repoPipeline.Execute(ctx)
	if err != nil {
		log.Errorf("Error while executing pipeline: %v", err)
		gitSetState(gitClient, "error", repoOwner, repoName, commitId)
		return
	}

	log.Debugf("Pipeline returned %d, setting state accordingly", exitcode)
	if exitcode == 0 {
		gitSetState(gitClient, "success", repoOwner, repoName, commitId)
	} else {
		gitSetState(gitClient, "failure", repoOwner, repoName, commitId)
	}
}

func getRepository(httpClient *http.Client, gitClient *github.Client, repoName string, repoOwner string,
	commitId string, repoToken *oauth2.Token) (string, error) {
	const GITHUB_ARCHIVE_URL = "https://github.com/%s/%s/archive/%s.zip"
	const GITHUB_ARCHIVE_ROOTDIR = "%s-%s-%s"
	const GITHUB_ARCHIVE_FORMAT = "zipball"
	var archiveUrl *url.URL
	var err error

	log.Info("Downloading archive of latest commit in push")
	if repoToken == nil {
		// no repoToken, so this is a public repository
		archiveUrl, err = url.Parse(fmt.Sprintf(GITHUB_ARCHIVE_URL, repoOwner, repoName, commitId))
		if err != nil {
			return "", errors.New(fmt.Sprintf("Error while constructing archive URL: %v", err))
		}
	} else {
		log.Debug("Getting archive URL for \"" + repoOwner + "/" + repoName + "\", ref \"" + commitId + "\"")
		archiveUrl, _, err = gitClient.Repositories.GetArchiveLink(repoOwner, repoName, GITHUB_ARCHIVE_FORMAT,
			&github.RepositoryContentGetOptions{Ref: commitId})
		if err != nil {
			return "", errors.New(fmt.Sprintf("Error while getting archive URL: %v", err))
		}
	}
	log.Debug("Found archive URL ", archiveUrl)

	tmpDir, err := ioutil.TempDir("", TMPDIR_PREFIX)
	if err != nil {
		log.Error()
		return "", errors.New(fmt.Sprintf("Error while creating temporary directory: %v", err))
	}

	log.Debug("Created temporary directory " + tmpDir)
	archivePath, err := downloadFile(httpClient, archiveUrl, tmpDir)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error while downloading archive: %v", err))
	}
	log.Debug("Archive downloaded to ", archivePath.Name())

	err = unzip(archivePath.Name(), tmpDir)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error while unpacking archive: %v", err))
	}

	// cleanup the archive
	archivePath.Close()
	os.Remove(archivePath.Name())

	// we should now have a copy of the repository at the latest commit
	repoDir := path.Join(tmpDir, fmt.Sprintf(GITHUB_ARCHIVE_ROOTDIR, repoOwner, repoName, commitId))
	if checkDirNotExists(repoDir) {
		repoDir = path.Join(tmpDir, fmt.Sprintf(GITHUB_ARCHIVE_ROOTDIR, repoOwner, repoName, commitId[0:7]))
		if checkDirNotExists(repoDir) {
			log.Error()
			return "", errors.New(fmt.Sprintf("Repository not found at expected directory \"%s\" after unpacking", repoDir))
		}
	}
	log.Debug("Repository unpacked to ", repoDir)

	return repoDir, nil
}

func downloadFile(httpClient *http.Client, url *url.URL, downloadDirectory string) (*os.File, error) {
	log.Debug("Downloading \"" + url.String() + "\"")
	filePath, err := ioutil.TempFile(downloadDirectory, TMPFILE_PREFIX)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Get(url.String())
	defer resp.Body.Close()
	n, err := io.Copy(filePath, resp.Body)
	if err != nil {
		return nil, err
	} else {
		log.Debugf("Downloaded %d bytes", n)
		return filePath, nil
	}
}

func readPipelineConfig(directory string) (pipeline.Pipeline, error) {
	var pipelineConfig pipeline.Pipeline
	pipelineConfigPath := path.Join(directory, strings.Join([]string{PIPELINEFILE, ".yaml"}, ""))
	if _, err := os.Stat(pipelineConfigPath); os.IsNotExist(err) == true {
		pipelineConfigPath = path.Join(directory, strings.Join([]string{PIPELINEFILE, ".yml"}, ""))
		if _, err := os.Stat(pipelineConfigPath); os.IsNotExist(err) == true {
			return pipelineConfig, errors.New("Couldn't find .octorunner.yaml or .octorunner.yml in repository")
		}
	}
	pipelineConfigBuf, err := ioutil.ReadFile(pipelineConfigPath)
	if err != nil {
		return pipelineConfig, errors.New(fmt.Sprintf("Error while reading from %s: %v", pipelineConfigPath, err))
	}

	pipelineConfig, err = pipeline.ParseConfig(pipelineConfigBuf)
	return pipelineConfig, err
}

func checkDirNotExists(dir string) bool {
	s, err := os.Stat(dir)
	return os.IsNotExist(err) == true || !s.IsDir()
}

func gitSetState(git *github.Client, state string, owner string, repo string, commit string) {
	git.Repositories.CreateStatus(owner, repo, commit, &github.RepoStatus{State: &state})
}

// Extract a zip file to a destination.
// By "swtdrgn" from http://stackoverflow.com/a/24430720
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				log.Error(err)
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
