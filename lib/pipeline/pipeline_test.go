package pipeline

import (
	"testing"
	"github.com/docker/docker/pkg/testutil/assert"
	"fmt"
)

const foo = `
image: test-foo

jobs:
  -
    script:
      - foo
    stage: build
    allow_failure: true
  -
    script:
      - do_test
    stage: test
    allow_failure: false
`

func TestConfigParsing(t *testing.T) {
	config, err := ParseConfig([]byte(foo))
	if err != nil {
		fmt.Printf("Error: %v", err)
		t.Fail()
	}

	fmt.Printf("%+v", config)
	assert.Equal(t, len(config.Jobs), 2)
}
