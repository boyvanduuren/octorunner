package pipeline

import (
	"golang.org/x/net/context"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

type ImageLister interface {
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
}

/*
 A pipeline contains an image name, and a few jobs that need to run on that image.
 For now, every job will be run on a container that uses the Pipeline's image.
*/
type Pipeline struct {
	Script []string `yaml:"script"`
	Image  string   `yaml:"image"`
}

/*
 Take a byte array and return a Pipeline.
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
	WORKDIR = "/var/run/octorunner"
)

/*
 Execute a given pipeline's first job.
 Todo: Run all the jobs
*/
func (c Pipeline) Execute(ctx context.Context) (int, error) {
	log.Info("Starting execution of pipeline")

	repoData, ok := ctx.Value("repoData").(map[string]string)
	if !ok {
		return -1, errors.New("Error while reading context")
	}

	// start image in described in job, execute script
	client, err := client.NewEnvClient()
	if err != nil {
		return -1, err
	}
	defer client.Close()

	imageFound, err := imageExists(ctx, client, c.Image)

	if !imageFound {
		// todo: enable registry authorization
		log.Infof("Pulling image \"%s\"", c.Image)
		reader, err := client.ImagePull(ctx, c.Image, types.ImagePullOptions{})
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
	container, err := client.ContainerCreate(ctx,
		&container.Config{
			Image:      c.Image,
			Entrypoint: strslice.StrSlice{"/bin/sh", "-c", commands},
			WorkingDir: WORKDIR},
		&container.HostConfig{
			AutoRemove: false,
			Binds:      []string{volumeBind}},
		&network.NetworkingConfig{},
		"test")
	if err != nil {
		return -1, err
	}
	// log warnings if we have some
	if len(container.Warnings) > 0 {
		log.Warnf("Received warnings while creating container: %v", container.Warnings)
	}
	log.Debugf("Container with ID \"%s\" created", container.ID)

	// start the container
	err = client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		return -1, err
	}

	// wait until the container is done
	client.ContainerWait(ctx, container.ID)

	// inspect the finished container so we can get the exitcode
	inspectData, err := client.ContainerInspect(ctx, container.ID)
	if err != nil {
		return -1, err
	}
	log.Infof("Container \"%s\" done, exit code: %d", container.ID, inspectData.State.ExitCode)

	log.Debugf("Removing container \"%s\"", container.ID)
	err = client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return -1, err
	}

	return inspectData.State.ExitCode, nil
}

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