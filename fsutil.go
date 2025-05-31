package main

import (
	"os"
	"path/filepath"
)

func WriteTempFile(pattern string, data []byte) (path string, err error) {
	dir, pattern := filepath.Split(pattern)

	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}

	name := f.Name()
	defer func() {
		if err != nil {
			os.Remove(name)
		}
	}()

	_, err = f.Write(data)
	if err != nil {
		f.Close()
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	return name, nil
}

func MkdirTemp(pattern string) (path string, err error) {
	dir, pattern := filepath.Split(pattern)
	return os.MkdirTemp(dir, pattern)
}
