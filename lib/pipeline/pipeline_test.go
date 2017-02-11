package pipeline

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
	"testing"
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
	expectedInt := 2
	if len(config.Script) != expectedInt {
		t.Fatalf("Expected config.Script to contain %d elements", expectedInt)
	}
	// the commands should be "true" and "true"
	expectedString := "true"
	if config.Script[0] != expectedString {
		t.Fatalf("Expected config.Script[0] to be \"%s\"", expectedString)
	}
	if config.Script[1] != "true" {
		t.Fatalf("Expected config.Script[1] to be \"%s\"", expectedString)
	}
	// image should be "alpine:latest"
	expectedString = "alpine:latest"
	if config.Image != expectedString {
		t.Fatalf("Expected config.Image to be \"%s\"", expectedString)
	}
}

// todo: this test now depends on a working docker host set in the environment, we need to mock this
//func TestPipelineExecute(t *testing.T) {
//	ctx := context.TODO()
//
//	ctx = context.WithValue(ctx, "repoData", map[string]string{
//		"fullName":   "boyvanduuren/octorunner",
//		"commitId":   "deadbeef",
//		"fsLocation": "/tmp/extracted",
//	})
//
//	config, _ := ParseConfig([]byte(foo))
//	ret, err := config.Execute(ctx)
//	if err != nil {
//		t.Fatalf("Expected no error, got: %v", err)
//	}
//	if ret != 0 {
//		t.Fatalf("Expected 0 from config.Execute(ctx), got %d", ret)
//	}
//}

type MockDockerClient struct {
	Images []string
	Err    error
}

func (client MockDockerClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	if client.Err != nil {
		return nil, client.Err
	} else {
		return []types.ImageSummary{{RepoTags: client.Images}}, nil
	}
}

func TestImageExists(t *testing.T) {
	client := MockDockerClient{Images: []string{"alpine:latest"}, Err: nil}
	exists, err := imageExists(context.TODO(), client, "alpine")
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	if exists == false {
		t.Fatal("Expected to find alpine, but didn't")
	}

	client.Err = errors.New("Testing an error")
	_, err = imageExists(context.TODO(), client, "alpine")
	if err == nil {
		t.Fatalf("Expected error: %v", client.Err)
	}
}

func TestContainerName(t *testing.T) {
	actual := containerName("boyvanduuren/octorunner", "deadbeef")
	expected := "boyvanduuren_octorunner-deadbeef"
	if actual != expected {
		t.Fatalf("Expected %s, but got %s", expected, actual)
	}

	actual = containerName("t#st!ng_some.STUFF-()", "cafebabe")
	expected = "tstng_some.STUFF--cafebabe"
	if actual != expected {
		t.Fatalf("Expected %s, but got %s", expected, actual)
	}
}
