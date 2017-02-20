package common

import (
	"os"
	"github.com/docker/docker/pkg/archive"
	"fmt"
	"io"
	"path"
	"path/filepath"
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