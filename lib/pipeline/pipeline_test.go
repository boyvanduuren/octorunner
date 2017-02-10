package pipeline

import (
	"fmt"
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
	config, err := ParseConfig([]byte(foo))
	if err != nil {
		fmt.Printf("Error: %v", err)
		t.Fail()
	}

	// we should have two lines in our script
	if len(config.Script) != 2 {
		t.Fatalf("Expected config.Script to contain 2 elements")
	}
	// the commands should be "true" and "true"
	if config.Script[0] != "true" {
		t.Fatalf("Expected config.Script[0] to be \"true\"")
	}
	if config.Script[1] != "true" {
		t.Fatalf("Expected config.Script[0] to be \"true\"")
	}
	// image should be "alpine:latest"
	if config.Image != "alpine:latest" {
		t.Fatalf("Expected config.Image to be \"alpine:latest\"")
	}
}

func TestPipelineExecute(t *testing.T) {
	// todo: this test now depends on a working docker host set in the environment, we need to mock this
	ctx := context.TODO()

	ctx = context.WithValue(ctx, "repoData", map[string]string{
		"fullName":   "boyvanduuren/octorunner",
		"commitId":   "deadbeef",
		"fsLocation": "/tmp/extracted",
	})

	config, _ := ParseConfig([]byte(foo))
	ret, err := config.Execute(ctx)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if ret != 0 {
		t.Fatalf("Expected 0 from config.Execute(ctx), got %d", ret)
	}
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