package main

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

func ReadModFile(path string) (*modfile.File, error) {
	if !strings.HasSuffix(path, ".mod") {
		path = filepath.Join(path, "go.mod")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return modfile.Parse(path, data, nil)
}
