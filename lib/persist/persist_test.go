package persist

import (
	"testing"
	"fmt"
)

func TestWriteOutput(t *testing.T) {
	OpenDatabase("octorunner.db")
	createProject("projectName", "projectOwner")
	id, err := findProjectID("projectName", "projectOwner");
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(id)
}
