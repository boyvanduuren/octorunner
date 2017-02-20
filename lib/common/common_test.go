package common

import (
	"testing"
	"os"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
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