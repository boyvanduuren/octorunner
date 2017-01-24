package pipeline

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	PIPELINEFILE = ".octorunner"
)

/*
 A job is one of the core structs in a pipeline configuration, it defines
 a unit of work. Every job belongs to a stage, and all jobs of a stage are run in parallel.
*/
type Job struct {
	Script       []string `yaml:"script"`
	Image        string   `yaml:"image"`
	Stage        string   `yaml:"stage"`
	AllowFailure bool     `yaml:"allow_failure"`
}

type PipelineConfig struct {
	Jobs   []Job    `yaml:"jobs"`
	Image  string   `yaml:"image"`
	Stages []string `yaml:"stages"`
}

func ParseConfig(file []byte) (PipelineConfig, error) {
	var pipelineConfig PipelineConfig
	err := yaml.Unmarshal(file, &pipelineConfig)
	if err != nil {
		return pipelineConfig, err
	}

	return pipelineConfig, nil
}

func ReadPipelineConfig(directory string) (PipelineConfig, error) {
	var pipelineConfig PipelineConfig
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

	pipelineConfig, err = ParseConfig(pipelineConfigBuf)
	return pipelineConfig, err
}
