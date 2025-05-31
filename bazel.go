package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Bazel struct {
	info map[string]string
}

func NewBazel(bazel, workspace string) (*Bazel, error) {
	workspace = filepath.Clean(workspace)

	cache := filepath.Join(os.TempDir(), fmt.Sprintf(
		"gopackagesdriver.%x.bazelinfo",
		md5.Sum([]byte(workspace)),
	))

	data, err := os.ReadFile(cache)
	if err != nil {
		cmd := exec.Command(bazel, "info",
			"bazel-genfiles",
			"workspace",
		)
		cmd.Dir = workspace

		data, err = cmd.Output()
		if err != nil {
			return nil, err
		}

		temp, err := WriteTempFile(cache, data)
		if err == nil {
			os.Rename(temp, cache)
		}
	}

	pairs := strings.Split(
		string(data), "\n",
	)

	info := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		key, value, found := strings.Cut(pair, ":")
		if found {
			info[key] = strings.TrimSpace(value)
		}
	}

	return &Bazel{
		info: info,
	}, nil
}

func (b *Bazel) Workspace() string {
	return b.info["workspace"]
}

func (b *Bazel) CacheDir() string {
	return b.info["bazel-genfiles"]
}

func (b *Bazel) WalkCacheDir(fn func(string, string) error) error {
	root := b.CacheDir()

	return filepath.WalkDir(root,
		func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				switch {
				case errors.Is(err, fs.ErrNotExist):
					return fs.SkipDir
				}

				return err
			}

			base, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			if !entry.IsDir() {
				return fn(path, base)
			}

			if base == "external" {
				return fs.SkipDir
			}

			if strings.HasSuffix(base, "_test_") {
				return fs.SkipDir
			}

			return nil
		},
	)
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

	path, err := filepath.Rel(prefix, path)
	if err != nil {
		panic(err)
	}

	return path, true
}
