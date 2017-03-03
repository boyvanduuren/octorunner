package persist

import (
	"github.com/Sirupsen/logrus"
	"testing"
)

func TestWriteOutput(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	OpenDatabase("octorunner.db")
	writer, err := CreateOutputWriter("foo", "bar", "deafbeef", "default")
	if err != nil {
		t.Fatal(err)
	}
	outputID, err := writer("woot")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Created output line with id %d", outputID)
}
