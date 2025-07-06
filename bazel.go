package main

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fxrlv/gopackagesdriver/internal/filecache"
)

func LookCacheDir(fsys *filecache.DirFS, bazel, workspace string) (string, error) {
	const name = "bazel-genfiles"

	data, err := fsys.ReadFile(name)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		cmd := exec.Command(bazel, "info", name)
		cmd.Dir = workspace

		data, err = cmd.Output()
		if err == nil {
			err = fsys.WriteFile(name, data)
		}
	}

	return strings.TrimSpace(string(data)), err
}

func WalkCacheDir(root string, fn func(string, string) error) error {
	_, err := os.Lstat(root)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			return nil
		}

		return err
	}

	walk := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			return fn(path, rel)
		}

		if rel == "external" {
			return fs.SkipDir
		}

		if strings.HasSuffix(rel, "_test_") {
			return fs.SkipDir
		}

		return nil
	}

	return filepath.WalkDir(root, walk)
}

func CutProtoPrefix(path string) (string, bool) {
	prefix := path
	for !strings.HasSuffix(prefix, "_go_proto_") {
		next := filepath.Dir(prefix)
		if next == prefix {
			return path, false
		}

		prefix = next
	}

	rel, err := filepath.Rel(prefix, path)
	if err != nil {
		panic(err)
	}

	return rel, true
}
