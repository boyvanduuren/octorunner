package pipeline

import (
	"bufio"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/boyvanduuren/octorunner/lib/common"
	"github.com/boyvanduuren/octorunner/lib/persist"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

/*
ImageLister implementations can be used to list available images on a Docker host.
*/
type ImageLister interface {
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
}

/*
ImagePuller implementations are used to pull Docker images to a Docker host.
*/
type ImagePuller interface {
	ImagePull(ctx context.Context, imageName string, options types.ImagePullOptions) (io.ReadCloser, error)
}

/*
ContainerCreater implementations are used to create containers on a Docker host.
*/
type ContainerCreater interface {
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
}

/*
ExecutionClient implementations are used to execute pipelines.
*/
type ExecutionClient interface {
	ImageLister
	ImagePuller
	ContainerCreater
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerWait(ctx context.Context, containerID string) (int64, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
}

/*
Pipeline contains an image name, and an array containing commands that are executed when
the pipeline is executed.
When the pipeline is executed, the script array will be concatenated as a single script, of which every
command needs to return 0 for the script to pass as successful.
*/
type Pipeline struct {
	Script []string `yaml:"script"`
	Image  string   `yaml:"image"`
}

const repositoryData string = "repositoryData"

/*
ParseConfig deserializes .octorunner.y[a]ml files contained in code repositories.
See https://github.com/boyvanduuren/octorunner#adding-a-test-to-your-repository.
*/
func ParseConfig(file []byte) (Pipeline, error) {
	var pipelineConfig Pipeline
	err := yaml.Unmarshal(file, &pipelineConfig)
	if err != nil {
		return pipelineConfig, err
	}

	return pipelineConfig, nil
}

// Extracted repositories are mounted as volumes on containers to WORKDIR.
const workDir = "/var/run/octorunner"

/*
Execute a pipeline, and return the exit code of its script.
*/
func (c Pipeline) Execute(ctx context.Context, cli ExecutionClient) (int, error) {
	log.Info("Starting execution of pipeline")

	repoData, ok := ctx.Value(repositoryData).(map[string]string)
	if !ok {
		return -1, errors.New("Error while reading context")
	}

	// look for image on Docker host, if we don't have it we'll pull it
	imageFound, err := imageExists(ctx, cli, c.Image)

	if !imageFound {
		log.Infof("Pulling image \"%s\"", c.Image)
		err := imagePull(ctx, cli, c.Image)
		if err != nil {
			return -1, err
		}
	} else {
		log.Debugf("Image \"%s\" is present", c.Image)
	}

	// create the container
	containerName := containerName(repoData["fullName"], repoData["commitId"])
	containerID, err := containerCreate(ctx, cli, c.Script, c.Image, containerName)
	if err != nil {
		return -1, err
	}

	// copy the working data to workDir
	log.Infof("Copying files from %q to container %q", repoData["fsLocation"], containerID)
	dst, src, out, err := common.CreateTarball(repoData["fsLocation"], workDir)
	if err != nil {
		return -1, fmt.Errorf("Error while preparing tarball: %q", err)
	}
	defer src.Close()
	defer out.Close()
	err = cli.CopyToContainer(ctx, containerID, dst, out, types.CopyToContainerOptions{AllowOverwriteDirWithFile: false})
	if err != nil {
		return -1, fmt.Errorf("Error while coping file(s): %q", err)
	}

	// start the container
	log.Infof("Starting container %q", containerID)
	err = cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return -1, fmt.Errorf("Error while starting container: %q", err)
	}
	containerRunning := true

	// if we have a connection to a db, log output
	if persist.Connection != nil {
		// get a writer that writes to the Output table in our database
		repoOwner := strings.Split(repoData["fullName"], "/")[0]
		repoName := strings.Split(repoData["fullName"], "/")[1]
		commitID := repoData["commitId"]
		writer, err := persist.CreateOutputWriter(repoName, repoOwner, commitID, "default", persist.Connection)
		if err != nil {
			return -1, err
		}
		go logOutput(ctx, cli, containerID, writer, &containerRunning)
	}

	// start a goroutine that logs output from the container

	// block until the container is done
	cli.ContainerWait(ctx, containerID)
	// signal the logOutput goroutine to stop
	containerRunning = false

	// inspect the finished container so we can get the exitcode
	inspectData, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return -1, fmt.Errorf("Error while inspecting container: %q", err)
	}
	log.Infof("Container \"%s\" done, exit code: %d", containerID, inspectData.State.ExitCode)

	log.Debugf("Removing container \"%s\"", containerID)
	err = cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return -1, fmt.Errorf("Error while removing container: %q", err)
	}

	return inspectData.State.ExitCode, nil
}

/*
Check whether a particular image is available on a Docker host. We need this information to
decide whether or not to pull the image.
*/
func imageExists(ctx context.Context, client ImageLister, imageName string) (bool, error) {
	// check if image exists
	log.Debugf("Looking if image \"%s\" is present on docker host", imageName)
	imageFound := false
	list, err := client.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return imageFound, err
	}
	for _, summary := range list {
		for _, tag := range summary.RepoTags {
			if imageName == tag || imageName == strings.Split(tag, ":")[0] {
				imageFound = true
			}
		}
	}

	return imageFound, nil
}

/*
Pull an image to a Docker host so it can be used to create containers.
*/
func imagePull(ctx context.Context, cli ImagePuller, imageName string) error {
	reader, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("Error while pulling %q: %q", imageName, err)
	}
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Error while pulling %q: %q", imageName, err)
	}
	log.Debugf("%s", buf)

	return nil
}

/*
Create a container using imageName on a Docker host with the given commands passed to "/bin/sh" as entrypoint.
Return the ID assigned to the container by Docker, or an error if something goes wrong.
*/
func containerCreate(ctx context.Context, cli ContainerCreater, commands []string, imageName string,
	containerName string) (string, error) {
	// create the container
	script := strings.Join(commands, " && ")
	log.Debugf("Creating container with entrypoint %q", script)
	container, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:      imageName,
			Entrypoint: strslice.StrSlice{"/bin/sh", "-c", script},
			WorkingDir: workDir},
		&container.HostConfig{AutoRemove: false},
		&network.NetworkingConfig{},
		containerName)

	if err != nil {
		return "", fmt.Errorf("Error while creating container: %q", err)
	}
	// log warnings if we have some
	if len(container.Warnings) > 0 {
		warnings := make([]string, len(container.Warnings))
		for i, s := range container.Warnings {
			warnings[i] = fmt.Sprintf("%q", s)
		}
		log.Warnf("Received warnings while creating container: %v", warnings)
	}
	log.Debugf("Container with ID %q created", container.ID)

	return container.ID, nil
}

/*
LogOutput checks a running container for log messages and passes messages with timestamps to a writer.
*/
func logOutput(ctx context.Context, cli ExecutionClient, containerID string,
	writer func(string, string) (int64, error), containerRunning *bool) error {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
	}

	rc, err := cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return err
	}
	defer rc.Close()

	scanner := bufio.NewScanner(rc)
	// while the container is running check for new log messages
	for *containerRunning {
		if scanner.Scan() {
			line := scanner.Text()
			// extract the message and date from the log message
			date, data, err := common.ExtractDateAndOutput(line)
			// we might have received an empty line, in which case we want to continue to the next iteration
			if err != nil {
				continue
			}
			// write the message and date to our writer
			_, err = writer(data, date)
			if err != nil {
				log.Errorf("Error while writing log to writer: %v", err)
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			log.Errorf("Error while scanning for log messages: %v")
			return err
		}
	}

	return nil
}

/*
Container names need to match [a-zA-Z_.-], so filter out everything that doesn't match.
Except "-", which is translated to "_".
*/
func containerName(repoFullName string, commitID string) string {
	mapping := func(r rune) rune {
		// Pattern compilation won't fail, so don't check for err
		match, _ := regexp.Match("[a-zA-Z_.-]", []byte{byte(r)})
		if match == false {
			if string(r) == "/" {
				return []rune("_")[0]
			}
			return -1
		}
		return r
	}

	return strings.Join([]string{strings.Map(mapping, repoFullName), commitID}, "-")
}
