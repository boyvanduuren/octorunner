package pipeline

import (
	"context"
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

/*
 A job is one of the core structs in a pipeline configuration, it defines
 a unit of work. Every job belongs to a stage, and all jobs of a stage are run in parallel.
*/
type Job struct {
	Script       []string `yaml:"script"`
	Image        string   `yaml:"image,omitempty"`
	AllowFailure bool     `yaml:"allow_failure"`
}

/*
 A pipeline contains an image name, and a few jobs that need to run on that image.
 For now, every job will be run on a container that uses the Pipeline's image.
*/
type Pipeline struct {
	Jobs  []Job  `yaml:"jobs"`
	Image string `yaml:"image"`
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

/*
 Execute a given pipeline's first job.
 Todo: Run all the jobs
*/
func (c Pipeline) Execute(ctx context.Context) (int, error) {
	log.Info("Starting execution of pipeline")

	// start image in described in job, execute script
	client, err := client.NewEnvClient()
	if err != nil {
		return -1, err
	}
	defer client.Close()

	// check if image exists
	log.Debugf("Looking if image \"%s\" is present on docker host", c.Image)
	list, err := client.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return -1, err
	}
	imageFound := false
	for _, summary := range list {
		for _, tag := range summary.RepoTags {
			if c.Image == tag || c.Image == strings.Split(tag, ":")[0] {
				imageFound = true
			}
		}
	}

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
	commands := strings.Join(c.Jobs[0].Script, " && ")
	log.Debugf("Creating container with entrypoint \"%s\"", commands)
	container, err := client.ContainerCreate(ctx,
		&container.Config{
			Image:      c.Image,
			Entrypoint: strslice.StrSlice{"/bin/sh", "-c", commands}},
		&container.HostConfig{AutoRemove: false},
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
