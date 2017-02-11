package pipeline

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
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
func (c Pipeline) Execute(ctx context.Context, cli *client.Client) (int, error) {
	log.Info("Starting execution of pipeline")

	repoData, ok := ctx.Value("repoData").(map[string]string)
	if !ok {
		return -1, errors.New("Error while reading context")
	}

	// start image in described in job, execute script
	imageFound, err := imageExists(ctx, cli, c.Image)

	if !imageFound {
		// todo: enable registry authorization
		log.Infof("Pulling image \"%s\"", c.Image)
		reader, err := cli.ImagePull(ctx, c.Image, types.ImagePullOptions{})
		if err != nil {
			return -1, err
		}
		buf, err := ioutil.ReadAll(reader)
		if err != nil {
			return -1, err
		}
		log.Debugf("%s", buf)
	} else {
		log.Debugf("Image \"%s\" is present", c.Image)
	}

	// create the container
	commands := strings.Join(c.Script, " && ")
	volumeBind := strings.Join([]string{repoData["fsLocation"], WORKDIR}, ":")
	log.Debugf("Creating container with entrypoint \"%s\" and bound volume \"%s\"", commands, volumeBind)
	container, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image:      c.Image,
			Entrypoint: strslice.StrSlice{"/bin/sh", "-c", commands},
			WorkingDir: WORKDIR},
		&container.HostConfig{
			AutoRemove: false,
			Binds:      []string{volumeBind}},
		&network.NetworkingConfig{},
		containerName(repoData["fullName"], repoData["commitId"]))
	if err != nil {
		return -1, err
	}
	// log warnings if we have some
	if len(container.Warnings) > 0 {
		log.Warnf("Received warnings while creating container: %v", container.Warnings)
	}
	log.Debugf("Container with ID \"%s\" created", container.ID)

	// start the container
	err = cli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		return -1, err
	}

	// wait until the container is done
	cli.ContainerWait(ctx, container.ID)

	// inspect the finished container so we can get the exitcode
	inspectData, err := cli.ContainerInspect(ctx, container.ID)
	if err != nil {
		return -1, err
	}
	log.Infof("Container \"%s\" done, exit code: %d", container.ID, inspectData.State.ExitCode)

	log.Debugf("Removing container \"%s\"", container.ID)
	err = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return -1, err
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
 Container names need to match [a-zA-Z_.-], so filter out everything that doesn't match.
 Except "-", which is translated to "_".
*/
func containerName(repoFullName string, commitId string) string {
	mapping := func(r rune) rune {
		match, err := regexp.Match("[a-zA-Z_.-]", []byte{byte(r)})
		if err != nil {
			return -1
		}
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
