package persist

import (
	"testing"
	"github.com/Sirupsen/logrus"
)

func TestWriteOutput(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	OpenDatabase("octorunner.db")
	err := WriteOutput("foo", "bar", "deafbeef", "default", "everything is fine")
	if err != nil {
		t.Fatal(err)
	}
}
