package pipeline

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
	"reflect"
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
	cases := []struct {
		c             MockDockerClient
		image         string
		expectedValue bool
		expectedErr   error
	}{
		{
			c: MockDockerClient{
				Images: []string{
					"alpine:latest",
				},
				Err: nil,
			},
			image:         "alpine",
			expectedValue: true,
			expectedErr:   nil,
		},
		{
			c: MockDockerClient{
				Images: []string{
					"alpine:latest",
					"golang:1.7.5",
				},
				Err: nil,
			},
			image:         "golang:1.7.5",
			expectedValue: true,
			expectedErr:   nil,
		},
		{
			c: MockDockerClient{
				Images: []string{
					"alpine:latest",
					"golang:1.7.5",
				},
				Err: nil,
			},
			image:         "debian",
			expectedValue: false,
			expectedErr:   nil,
		},
		{
			c: MockDockerClient{
				Images: []string{},
				Err:    errors.New("Client error"),
			},
			image:         "doesn't_matter",
			expectedValue: false,
			expectedErr:   errors.New("Client error"),
		},
	}
	for _, testCase := range cases {
		val, err := imageExists(context.TODO(), testCase.c, testCase.image)
		if !reflect.DeepEqual(err, testCase.expectedErr) {
			t.Errorf("Expected err to be %q,but it was %q", testCase.expectedErr, err)
		}

		if testCase.expectedValue != val {
			t.Errorf("Expected %q, but got %q", testCase.expectedValue, val)
		}
	}
}

func TestContainerName(t *testing.T) {
	cases := []struct {
		repoFullName  string
		commitId      string
		expectedValue string
	}{
		{
			repoFullName:  "boyvanduuren/octorunner",
			commitId:      "deadbeef",
			expectedValue: "boyvanduuren_octorunner-deadbeef",
		},
		{
			repoFullName:  "t#st!ng_some.STUFF-()",
			commitId:      "cafebabe",
			expectedValue: "tstng_some.STUFF--cafebabe",
		},
	}

	for _, testCase := range cases {
		val := containerName(testCase.repoFullName, testCase.commitId)
		if testCase.expectedValue != val {
			t.Errorf("Expected %s, but got %s", testCase.expectedValue, val)
		}
	}
}
