package common

import (
	"archive/zip"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

/*
CheckDirNotExists returns true if a given path doesn't exist, or is not a directory.
*/
func CheckDirNotExists(dir string) bool {
	s, err := os.Stat(dir)
	return os.IsNotExist(err) == true || !s.IsDir()
}

/*
CreateTarball takes a source and destination. A string will be returned denoting the path the tarball should
be copied to for the target container. The first ReadCloser is opened on the source path, the second represents
the tarball in memory.
*/
func CreateTarball(source string, destination string) (string, io.ReadCloser, io.ReadCloser, error) {
	if CheckDirNotExists(source) {
		return "", nil, nil, fmt.Errorf("%q does not exist", source)
	}
	if !path.IsAbs(destination) {
		return "", nil, nil, fmt.Errorf("%q is not absolute", destination)
	}

	srcInfo, err := archive.CopyInfoSourcePath(source, false)
	if err != nil {
		return "", nil, nil, err
	}

	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return "", nil, nil, err
	}

	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, archive.CopyInfo{Path: destination})
	if err != nil {
		return "", nil, nil, err
	}

	return filepath.ToSlash(dstDir), srcArchive, preparedArchive, nil
}

/*
Unzip a zip archive to dest. Return the first directory unzipped, because in the case of a zip archive downloaded from
Github, that will always be the root of the repository.
Based on code by "swtdrgn" from http://stackoverflow.com/a/24430720
*/
func Unzip(src, dest string) (string, error) {
	var firstEncounteredDir string
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			// If this is the first dir we encounter, it's the root directory. Assign it so we can return it.
			if firstEncounteredDir == "" {
				firstEncounteredDir = fpath
			}
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				return "", err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return "", err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return "", err
			}
		}
	}
	return firstEncounteredDir, nil
}

/*
ExtractDateAndOutput takes a log line from a Docker container and extracts the RFC3339 timestamp and log message.
*/
func ExtractDateAndOutput(s string) (string, string, error) {
	pattern := regexp.MustCompile("^.*([0-9]{4}-[0-1][0-9]-[0-3][0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9]\\." +
		"[0-9]{9}Z) (.*)$")

	if pattern.Match([]byte(s)) {
		matches := pattern.FindStringSubmatch(s)
		return matches[1], matches[2], nil
	}
	return "", "", fmt.Errorf("Unable to extract time or data from %q", s)
}
