package pipeline

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/testutil/assert"
	"testing"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

const foo = `
image: alpine:latest
script:
  - true
  - true
`

func TestConfigParsing(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ParseConfig([]byte(foo))
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

	config, _ := ParseConfig([]byte(foo))
	ret, err := config.Execute(ctx)
	assert.NilError(t, err)
	assert.Equal(t, ret, 0)
}

type MockDockerClient struct {
	Images []string
	Err error
}

func (client MockDockerClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	if client.Err != nil {
		return nil, client.Err
	} else {
		return []types.ImageSummary{types.ImageSummary{RepoTags: client.Images}}, nil
	}
}

func TestImageExists(t *testing.T) {
	client := MockDockerClient{Images: []string{"alpine:latest"}, Err: nil}
	exists, err := imageExists(context.TODO(), client, "alpine")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v")
	}
	if exists == false {
		t.Fatalf("Expected to find alpine, but didn't")
	}
}