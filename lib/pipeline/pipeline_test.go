package pipeline_test

import (
	"context"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/boyvanduuren/octorunner/lib/pipeline"
	"github.com/docker/docker/pkg/testutil/assert"
	"testing"
)

const foo = `
image: alpine:latest
script:
  - true
  - true
`

func TestConfigParsing(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := pipeline.ParseConfig([]byte(foo))
	if err != nil {
		fmt.Printf("Error: %v", err)
		t.Fail()
	}

	// we should have two lines in our script
	assert.Equal(t, len(config.Script), 2)
	// the commands should be "true" and "true"
	assert.Equal(t, config.Script[0], "true")
	assert.Equal(t, config.Script[1], "true")
	// image should be "alpine:latest"
	assert.Equal(t, config.Image, "alpine:latest")
}

func TestPipelineExecute(t *testing.T) {
	// todo: this test now depends on a working docker host set in the environment, we need to mock this

	log.SetLevel(log.DebugLevel)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, "repoData", map[string]string{
		"fullName":   "boyvanduuren/octorunner",
		"commitId":   "deadbeef",
		"fsLocation": "/tmp/extracted",
	})

	config, _ := pipeline.ParseConfig([]byte(foo))
	ret, err := config.Execute(ctx)
	assert.NilError(t, err)
	assert.Equal(t, ret, 0)
}
