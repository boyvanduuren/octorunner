package pipeline

import (
	"gopkg.in/yaml.v2"
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

type Pipeline struct {
	Jobs   []Job    `yaml:"jobs"`
	Image  string   `yaml:"image"`
}

func ParseConfig(file []byte) (Pipeline, error) {
	var pipelineConfig Pipeline
	err := yaml.Unmarshal(file, &pipelineConfig)
	if err != nil {
		return pipelineConfig, err
	}

	return pipelineConfig, nil
}

func (c Pipeline) Execute() (int) {
	// start image in described in job, execute script

	return 0
}
