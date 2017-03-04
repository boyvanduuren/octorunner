package common

import (
	"testing"
	"os"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"fmt"
)

func TestCreateTarball(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(tempDir, "test"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	dst, src, out, err := CreateTarball(tempDir, "/var/run/octorunner/")
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	defer out.Close()

	if dst != "/var/run" {
		t.Fatalf("Expected destination %q", "/var/run")
	}

	outFile, err := os.Create(path.Join(tempDir, "out.tar"))
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()

	w, err := io.Copy(outFile, out)
	if err != nil {
		t.Fatal(err)
	}
	if w == 0 {
		t.Fatal("Expected more than one byte to be written")
	}
}

func TestExtractDateAndOutput(t *testing.T) {
	expectedDate := "2017-03-04T11:58:34.890790992Z"
	expectedData := "Foobar"
	date, data, err := ExtractDateAndOutput(fmt.Sprintf("\x00\x00%s %s", expectedDate, expectedData))
	if err != nil {
		t.Fatalf("Got unexpected error: %q", err)
	}

	if date != expectedDate {
		t.Fatalf("Date %q didn't match %q", date, expectedData)
	}
	if data != expectedData {
		t.Fatalf("Data didn't %q didn't match %q", data, expectedData)
	}
}