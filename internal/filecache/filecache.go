package filecache

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type DirFS struct {
	dir string
}

func Open(dir string, seed []byte) (*DirFS, error) {
	path := filepath.Join(dir, fmt.Sprintf(
		"gopackagesdriver-%x", md5.Sum(seed),
	))

	err := os.Mkdir(path, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, err
	}

	return &DirFS{
		dir: path,
	}, nil
}

func (fs *DirFS) Dir() string {
	return fs.dir
}

func (fs *DirFS) ReadFile(name string) ([]byte, error) {
	path := filepath.Join(fs.dir, name)
	return os.ReadFile(path)
}

func (fs *DirFS) WriteFile(name string, data []byte) error {
	f, err := os.CreateTemp(fs.dir, name+"-*")
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		f.Close()
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	path := filepath.Join(fs.dir, name)
	return os.Rename(f.Name(), path)
}
