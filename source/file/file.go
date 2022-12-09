package file

import (
	"github.com/dityaaa/concept/source"
	"io"
	nurl "net/url"
	"os"
	"path/filepath"
)

var _ source.Driver = (*File)(nil)

type Config struct {
	MigrationPath string
}

type File struct {
	migrationPath string

	files         []os.DirEntry
	count         int
	curIndex      int
	curIdentifier string
	curScript     io.ReadCloser
	curError      error
}

func Open(url string) (source.Driver, error) {
	purl, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	migrationPath := filepath.Join(purl.Host, purl.Path)

	files, err := os.ReadDir(migrationPath)
	if err != nil {
		return nil, err
	}

	return &File{
		migrationPath: migrationPath,
		files:         files,
		count:         len(files),
		curIndex:      0,
	}, nil
}

func (i *File) Name() string {
	return "file"
}

func (i *File) Close() error {
	return nil
}

func (i *File) Next() bool {
	if i.curIndex >= i.count {
		// set to nil because it's a memory hog
		i.files = nil
		return false
	}

	if i.files[i.curIndex].IsDir() {
		i.curIndex++
		return i.Next()
	}

	dirEntry := i.files[i.curIndex]
	file, err := os.Open(filepath.Join(i.migrationPath, dirEntry.Name()))
	if err != nil {
		i.curError = err
		return false
	}

	i.curIdentifier = file.Name()
	i.curScript = file
	i.curIndex++
	return true
}

func (i *File) Read() (*source.Migration, error) {
	return &source.Migration{
		Identifier: i.curIdentifier,
		Script:     i.curScript,
	}, nil
}

func (i *File) Touch(name string) error {
	_, err := os.Create(name)
	return err
}

func (i *File) Remove(name string) error {
	return os.Remove(name)
}

func (i *File) Err() error {
	return i.curError
}
