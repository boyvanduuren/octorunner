package pipeline

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
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

type PipelineExecutionClient interface {
	ImageLister
	ImagePuller
	ContainerCreater
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerWait(ctx context.Context, containerID string) (int64, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
}

/*
 A pipeline contains an image name, and an array containing commands that are executed when
 the pipeline is executed.
 When the pipeline is executed, the script array will be concatenated as a single script, of which every
 command needs to return 0 for the script to pass as successful.
*/
type Pipeline struct {
	Script []string `yaml:"script"`
	Image  string   `yaml:"image"`
}

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

const (
	// Extracted repositories are mounted as volumes on containers to WORKDIR.
	WORKDIR = "/var/run/octorunner"
)

/*
 Execute a pipeline, and return the exit code of its script.
*/
func (c Pipeline) Execute(ctx context.Context, cli PipelineExecutionClient) (int, error) {
	log.Info("Starting execution of pipeline")

	repoData, ok := ctx.Value("repoData").(map[string]string)
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

	// start the container
	volumeBind := strings.Join([]string{repoData["fsLocation"], WORKDIR}, ":")
	containerName := containerName(repoData["fullName"], repoData["commitId"])
	containerId, err := containerCreate(ctx, cli, c.Script, volumeBind, c.Image, containerName)
	if err != nil {
		return -1, err
	}

	err = cli.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error while starting container: %q", err))
	}

	// wait until the container is done
	cli.ContainerWait(ctx, containerId)

	// inspect the finished container so we can get the exitcode
	inspectData, err := cli.ContainerInspect(ctx, containerId)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error while inspecting container: %q", err))
	}
	log.Infof("Container \"%s\" done, exit code: %d", containerId, inspectData.State.ExitCode)

	log.Debugf("Removing container \"%s\"", containerId)
	err = cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return -1, errors.New(fmt.Sprintf("Error while removing container: %q", err))
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
		return errors.New(fmt.Sprintf("Error while pulling %q: %q", imageName, err))
	}
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while pulling %q: %q", imageName, err))
	}
	log.Debugf("%s", buf)

	return nil
}

/*
 Create a container using imageName on a Docker host with the given commands passed to "/bin/sh" as entrypoint, and
 bindDir mounted to WORKDIR (see Constants).
 Return the ID assigned to the container by Docker, or an error if something goes wrong.
*/
func containerCreate(ctx context.Context, cli ContainerCreater, commands []string, bindDir string, imageName string,
	containerName string) (string, error) {
	// create the container
	script := strings.Join(commands, " && ")
	log.Debugf("Creating container with entrypoint %q and bound volume %q", script, bindDir)
	container, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:      imageName,
			Entrypoint: strslice.StrSlice{"/bin/sh", "-c", script},
			WorkingDir: WORKDIR},
		&container.HostConfig{
			AutoRemove: false,
			Binds:      []string{bindDir}},
		&network.NetworkingConfig{},
		containerName)

	if err != nil {
		return "", errors.New(fmt.Sprintf("Error while creating container: %q", err))
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
 Container names need to match [a-zA-Z_.-], so filter out everything that doesn't match.
 Except "-", which is translated to "_".
*/
func containerName(repoFullName string, commitId string) string {
	mapping := func(r rune) rune {
		// Pattern compilation won't fail, so don't check for err
		match, _ := regexp.Match("[a-zA-Z_.-]", []byte{byte(r)})
		if match == false {
			if string(r) == "/" {
				return []rune("_")[0]
			} else {
				return -1
			}
		} else {
			return r
		}
	}

	return strings.Join([]string{strings.Map(mapping, repoFullName), commitId}, "-")
}
