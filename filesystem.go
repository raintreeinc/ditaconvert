package ditaconvert

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type FileSystem interface {
	ReadFile(path string) (data []byte, modified time.Time, err error)
}

func CanonicalPath(name string) string { return strings.ToLower(name) }
func trimext(name string) string       { return name[0 : len(name)-len(filepath.Ext(name))] }

type Dir string

func (dir Dir) fullpath(name string) string {
	return filepath.FromSlash(path.Join(string(dir), name))
}

func (dir Dir) ReadFile(name string) (data []byte, modified time.Time, err error) {
	var file *os.File

	file, err = os.Open(dir.fullpath(name))
	if err != nil {
		return
	}

	var stat os.FileInfo
	stat, err = file.Stat()
	if err != nil {
		return
	}
	modified = stat.ModTime()

	data, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}

	return
}

type VFS map[string]string

func (fs VFS) ReadFile(name string) (data []byte, modified time.Time, err error) {
	content, ok := fs[name]
	if !ok {
		return nil, time.Time{}, os.ErrNotExist
	}
	return []byte(content), time.Now(), nil
}
