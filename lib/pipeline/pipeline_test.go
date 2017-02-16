package pipeline

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestConfigParsing(t *testing.T) {
	yaml := `
image: alpine:latest
script:
  - true
  - true
`

	config, err := ParseConfig([]byte(yaml))
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

	yaml = "won't parse"
	_, err = ParseConfig([]byte(yaml))
	if err == nil {
		t.Fail()
	}
}

// todo: this test now depends on a working docker host set in the environment, we need to mock this
//func TestPipelineExecute(t *testing.T) {
//	ctx := context.TODO()
//
//	ctx = context.WithValue(ctx, git.RepositoryData, map[string]string{
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

type MockImageLister struct {
	Images []string
	Err    error
}

func (client MockImageLister) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	if client.Err != nil {
		return nil, client.Err
	}
	return []types.ImageSummary{{RepoTags: client.Images}}, nil
}

func TestImageExists(t *testing.T) {
	cases := []struct {
		c             MockImageLister
		image         string
		expectedValue bool
		expectedErr   error
	}{
		{
			c: MockImageLister{
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
			c: MockImageLister{
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
			c: MockImageLister{
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
			c: MockImageLister{
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
			t.Errorf("Expected %t, but got %t", testCase.expectedValue, val)
		}
	}
}

type MockImagePuller struct {
	Err error
	Out io.ReadCloser
}

// Used to test the error routine of imagePull.
type ErrorReadCloser struct{}

func (errorReadCloser ErrorReadCloser) Read(p []byte) (int, error) {
	return 0, errors.New("This always errors")
}
func (errorReadCloser ErrorReadCloser) Close() error {
	return errors.New("This always errors")
}

func (client MockImagePuller) ImagePull(ctx context.Context, imageName string, options types.ImagePullOptions) (io.ReadCloser, error) {
	if client.Err != nil {
		return nil, client.Err
	}
	return client.Out, nil
}

func TestImagePull(t *testing.T) {
	cases := []struct {
		c             ImagePuller
		imageName     string
		expectedError error
	}{
		{
			c: MockImagePuller{
				Err: errors.New("No such image exists"),
				Out: ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
			},
			imageName:     "foobar",
			expectedError: fmt.Errorf("Error while pulling %q: %q", "foobar", "No such image exists"),
		},
		{
			c: MockImagePuller{
				Err: nil,
				Out: ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
			},
			imageName:     "golang:latest",
			expectedError: nil,
		},
		{
			c: MockImagePuller{
				Err: nil,
				Out: ErrorReadCloser{},
			},
			imageName:     "golang:latest",
			expectedError: fmt.Errorf("Error while pulling %q: %q", "golang:latest", "This always errors"),
		},
	}

	for _, testCase := range cases {
		err := imagePull(context.TODO(), testCase.c, testCase.imageName)
		if !reflect.DeepEqual(err, testCase.expectedError) {
			t.Errorf("Expected err to be %q, but it was %q", testCase.expectedError, err)
		}
	}
}

type MockContainerCreater struct {
	Warnings []string
	ID       string
	Err      error
}

func (client MockContainerCreater) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	if client.Err != nil {
		return container.ContainerCreateCreatedBody{}, client.Err
	}

	container := container.ContainerCreateCreatedBody{
		Warnings: client.Warnings,
		ID:       client.ID,
	}

	return container, nil
}

func TestContainerCreate(t *testing.T) {
	cases := []struct {
		c             ContainerCreater
		expectedValue string
		expectedError error
	}{
		{
			c: MockContainerCreater{
				ID: "createdId",
				Warnings: []string{
					"Some warning",
					"Ph'nglui mglw'nafh Cthulhu R'lyeh wgah'nagl fhtagn",
				},
			},
			expectedValue: "createdId",
		},
		{
			c: MockContainerCreater{
				ID:  "createdId",
				Err: errors.New("Container creation error"),
			},
			expectedValue: "",
			expectedError: fmt.Errorf("Error while creating container: %q", "Container creation error"),
		},
	}

	for _, testCase := range cases {
		val, err := containerCreate(context.TODO(), testCase.c, []string{"true"}, "golang:latest",
			"boyvanduuren_octorunner-1234")
		if !reflect.DeepEqual(err, testCase.expectedError) {
			t.Errorf("Expected err to be %q, but it was %q", testCase.expectedError, err)
		}
		if testCase.expectedValue != val {
			t.Errorf("Expected %q, but got %q", testCase.expectedValue, val)
		}
	}
}

type MockPipelineExecutionClient struct {
	ListErr    error
	ListImages []string
	PullErr    error
	PullOut    io.ReadCloser
	CreateErr  error
	CreateID   string
	StartErr   error
	WaitErr    error
	InspectErr error
	ExitCode   int
	RemoveErr  error
}

func (client MockPipelineExecutionClient) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	if client.ListErr != nil {
		return nil, client.ListErr
	}
	return []types.ImageSummary{{RepoTags: client.ListImages}}, nil
}

func (client MockPipelineExecutionClient) ImagePull(ctx context.Context, imageName string, options types.ImagePullOptions) (io.ReadCloser, error) {
	if client.PullErr != nil {
		return nil, client.PullErr
	}
	return client.PullOut, nil
}

func (client MockPipelineExecutionClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	if client.CreateErr != nil {
		return container.ContainerCreateCreatedBody{}, client.CreateErr
	}

	container := container.ContainerCreateCreatedBody{
		ID: client.CreateID,
	}

	return container, nil
}

func (client MockPipelineExecutionClient) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	return client.StartErr
}

func (client MockPipelineExecutionClient) ContainerWait(ctx context.Context, containerID string) (int64, error) {
	return 0, client.WaitErr
}

func (client MockPipelineExecutionClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if client.InspectErr == nil {
		return types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{
					ExitCode: client.ExitCode,
				},
			}}, nil
	}
	return types.ContainerJSON{}, client.InspectErr
}

func (client MockPipelineExecutionClient) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return client.RemoveErr
}

func (client MockPipelineExecutionClient) CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error {
	return nil
}

func TestPipelineExecute(t *testing.T) {
	cases := []struct {
		p             Pipeline
		c             MockPipelineExecutionClient
		ctx           context.Context
		expectedValue int
		expectedError error
	}{
		// Invalid context
		{
			p: Pipeline{
				Image: "golang:latest",
				Script: []string{
					"true",
				},
			},
			c: MockPipelineExecutionClient{
				ListErr: nil,
				ListImages: []string{
					"golang:latest",
				},
				PullErr:   nil,
				PullOut:   ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
				CreateErr: nil,
				CreateID:  "foo",
			},
			ctx:           context.TODO(),
			expectedValue: -1,
			expectedError: errors.New("Error while reading context"),
		},
		// Successful, image exists, no errors
		{
			p: Pipeline{
				Image: "golang:latest",
				Script: []string{
					"true",
				},
			},
			c: MockPipelineExecutionClient{
				ListErr: nil,
				ListImages: []string{
					"golang:latest",
				},
				PullErr:   nil,
				PullOut:   ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
				CreateErr: nil,
				CreateID:  "foo",
			},
			ctx: context.WithValue(context.TODO(), repositoryData, map[string]string{
				"fullName":   "boyvanduuren/octorunner",
				"fsLocation": "/tmp/",
				"commitId":   "deadbeef",
			}),
			expectedValue: 0,
			expectedError: nil,
		},
		// Image doesn't exist, but we can succesfully pull it
		{
			p: Pipeline{
				Image: "archlinux:latest",
				Script: []string{
					"true",
				},
			},
			c: MockPipelineExecutionClient{
				ListErr: nil,
				ListImages: []string{
					"golang:latest",
				},
				PullErr:   nil,
				PullOut:   ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
				CreateErr: nil,
				CreateID:  "foo",
			},
			ctx: context.WithValue(context.TODO(), repositoryData, map[string]string{
				"fullName":   "boyvanduuren/octorunner",
				"fsLocation": "/tmp/",
				"commitId":   "deadbeef",
			}),
			expectedValue: 0,
			expectedError: nil,
		},
		// Image doesn't exist, and we get an error while pulling
		{
			p: Pipeline{
				Image: "archlinux:latest",
				Script: []string{
					"true",
				},
			},
			c: MockPipelineExecutionClient{
				ListErr: nil,
				ListImages: []string{
					"golang:latest",
				},
				PullErr:   errors.New("It failed"),
				PullOut:   ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
				CreateErr: nil,
				CreateID:  "foo",
			},
			ctx: context.WithValue(context.TODO(), repositoryData, map[string]string{
				"fullName":   "boyvanduuren/octorunner",
				"fsLocation": "/tmp/",
				"commitId":   "deadbeef",
			}),
			expectedValue: -1,
			expectedError: fmt.Errorf("Error while pulling %q: %q", "archlinux:latest", "It failed"),
		},
		// Error while creating container
		{
			p: Pipeline{
				Image: "archlinux:latest",
				Script: []string{
					"true",
				},
			},
			c: MockPipelineExecutionClient{
				ListErr: nil,
				ListImages: []string{
					"golang:latest",
				},
				PullErr:   nil,
				PullOut:   ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
				CreateErr: errors.New("Creation error"),
				CreateID:  "foo",
			},
			ctx: context.WithValue(context.TODO(), repositoryData, map[string]string{
				"fullName":   "boyvanduuren/octorunner",
				"fsLocation": "/tmp/",
				"commitId":   "deadbeef",
			}),
			expectedValue: -1,
			expectedError: fmt.Errorf("Error while creating container: %q", "Creation error"),
		},
		// Error while starting container
		{
			p: Pipeline{
				Image: "archlinux:latest",
				Script: []string{
					"true",
				},
			},
			c: MockPipelineExecutionClient{
				ListErr: nil,
				ListImages: []string{
					"golang:latest",
				},
				PullOut:  ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
				CreateID: "foo",
				StartErr: errors.New("Start error"),
			},
			ctx: context.WithValue(context.TODO(), repositoryData, map[string]string{
				"fullName":   "boyvanduuren/octorunner",
				"fsLocation": "/tmp/",
				"commitId":   "deadbeef",
			}),
			expectedValue: -1,
			expectedError: fmt.Errorf("Error while starting container: %q", "Start error"),
		},
		// Error while inspecting container
		{
			p: Pipeline{
				Image: "archlinux:latest",
				Script: []string{
					"true",
				},
			},
			c: MockPipelineExecutionClient{
				ListErr: nil,
				ListImages: []string{
					"archlinux:latest",
				},
				PullOut:    ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
				CreateID:   "foo",
				InspectErr: errors.New("Inspection error"),
			},
			ctx: context.WithValue(context.TODO(), repositoryData, map[string]string{
				"fullName":   "boyvanduuren/octorunner",
				"fsLocation": "/tmp/",
				"commitId":   "deadbeef",
			}),
			expectedValue: -1,
			expectedError: fmt.Errorf("Error while inspecting container: %q", "Inspection error"),
		},
		// Error while removing container
		{
			p: Pipeline{
				Image: "archlinux:latest",
				Script: []string{
					"true",
				},
			},
			c: MockPipelineExecutionClient{
				ListErr: nil,
				ListImages: []string{
					"archlinux:latest",
				},
				PullOut:   ioutil.NopCloser(bytes.NewBufferString("doesn't matter")),
				CreateID:  "foo",
				RemoveErr: errors.New("Remove error"),
			},
			ctx: context.WithValue(context.TODO(), repositoryData, map[string]string{
				"fullName":   "boyvanduuren/octorunner",
				"fsLocation": "/tmp/",
				"commitId":   "deadbeef",
			}),
			expectedValue: -1,
			expectedError: fmt.Errorf("Error while removing container: %q", "Remove error"),
		},
	}

	for _, testCase := range cases {
		val, err := testCase.p.Execute(testCase.ctx, testCase.c)
		if !reflect.DeepEqual(err, testCase.expectedError) {
			t.Errorf("Expected err to be %q, but it was %q", testCase.expectedError, err)
		}
		if testCase.expectedValue != val {
			t.Errorf("Expected %d, but got %d", testCase.expectedValue, val)
		}
	}
}

func TestContainerName(t *testing.T) {
	cases := []struct {
		repoFullName  string
		commitID      string
		expectedValue string
	}{
		{
			repoFullName:  "boyvanduuren/octorunner",
			commitID:      "deadbeef",
			expectedValue: "boyvanduuren_octorunner-deadbeef",
		},
		{
			repoFullName:  "t#st!ng_some.STUFF-()",
			commitID:      "cafebabe",
			expectedValue: "tstng_some.STUFF--cafebabe",
		},
	}

	for _, testCase := range cases {
		val := containerName(testCase.repoFullName, testCase.commitID)
		if testCase.expectedValue != val {
			t.Errorf("Expected %s, but got %s", testCase.expectedValue, val)
		}
	}
}
