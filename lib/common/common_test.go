package common

import "testing"

func TestFindAllFiles(t *testing.T) {
	t.Log(FindAllFiles("C:\\Windows\\Temp"))
	t.Log(FindAllFiles("foo"))
}
