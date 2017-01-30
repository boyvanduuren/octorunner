package pipeline_test

import (
	"testing"
	"github.com/docker/docker/pkg/testutil/assert"
	"fmt"
	"github.com/boyvanduuren/octorunner/lib/pipeline"
)

const foo = `
image: test-foo

jobs:
  -
    script:
      - foo
      - bar
    allow_failure: true
  -
    script:
      - do_test
    image: test/image:latest
    allow_failure: false
`

func TestConfigParsing(t *testing.T) {
	config, err := pipeline.ParseConfig([]byte(foo))
	if err != nil {
		fmt.Printf("Error: %v", err)
		t.Fail()
	}

	// we should have two jobs
	assert.Equal(t, len(config.Jobs), 2)

	job1 := config.Jobs[0]
	// the first job should have two commands in the script
	assert.Equal(t, len(job1.Script), 2)
	// and they should be "foo" and "bar"
	assert.Equal(t, job1.Script[0], "foo");
	assert.Equal(t, job1.Script[1], "bar");
	// allow_failure should be true
	assert.Equal(t, job1.AllowFailure, true)
	// image should be empty
	assert.Equal(t, job1.Image, "")

	job2 := config.Jobs[1]
	// the second job should only have one command
	assert.Equal(t, len(job2.Script), 1)
	// it should equal "do_test"
	assert.Equal(t, job2.Script[0], "do_test")
	// image should be "test/image:latest"
	assert.Equal(t, job2.Image, "test/image:latest")
}
