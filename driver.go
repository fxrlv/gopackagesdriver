package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/fxrlv/gopackagesdriver/internal/filecache"
	"github.com/fxrlv/gopackagesdriver/internal/fsutil"
)

type DriverConfig struct {
	WorkspaceDir string
	WorkingDir   string
	Bazel        string
}

type Driver struct {
	workspaceDir string
	workingDir   string
	bazel        string

	fsys *filecache.DirFS
}

func NewDriver(config *DriverConfig) (*Driver, error) {
	if config == nil || config.WorkspaceDir == "" {
		return &Driver{}, nil
	}

	fsys, err := filecache.Open(os.TempDir(), []byte(config.WorkspaceDir))
	if err != nil {
		return nil, err
	}

	bazel := config.Bazel
	if bazel == "" {
		bazel = "bazel"
	}

	return &Driver{
		workspaceDir: config.WorkspaceDir,
		workingDir:   config.WorkingDir,
		bazel:        bazel,
		fsys:         fsys,
	}, nil
}

func (d *Driver) Serve(req *packages.DriverRequest, patterns []string) (*packages.DriverResponse, error) {
	if d.workspaceDir == "" {
		return &packages.DriverResponse{
			NotHandled: true,
		}, nil
	}

	rel, err := filepath.Rel(d.workspaceDir, d.workingDir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return &packages.DriverResponse{
			NotHandled: true,
		}, nil
	}

	if rel == "." {
		i := slices.Index(patterns, "./...")
		if i >= 0 {
			patterns[i] = patterns[len(patterns)-1]
			patterns = patterns[:len(patterns)-1]
		}
	}

	if len(patterns) == 0 {
		patterns = append(patterns, "builtin")
	}

	cfg := packages.Config{
		Mode:       req.Mode,
		Env:        req.Env,
		BuildFlags: req.BuildFlags,
		Tests:      req.Tests,
		Overlay:    req.Overlay,
		Dir:        d.workingDir,
	}

	return d.LoadWorkspace(&cfg, patterns)
}

func (d *Driver) LoadWorkspace(cfg *packages.Config, patterns []string) (*packages.DriverResponse, error) {
	overlay := fsutil.NewOverlay()

	err := d.ReadCacheDir(overlay)
	if err != nil {
		return nil, err
	}

	if overlay.Len() > 0 && len(cfg.Overlay) > 0 {
		dir, err := os.MkdirTemp(d.fsys.Dir(), "overlay-*")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(dir)

		err = overlay.Append(dir, cfg.Overlay)
		if err != nil {
			return nil, err
		}

		cfg.Overlay = nil
	}

	if overlay.Len() > 0 {
		path, err := WriteOverlay(d.fsys.Dir(), overlay)
		if err != nil {
			return nil, err
		}
		defer os.Remove(path)

		cfg.BuildFlags = append(cfg.BuildFlags,
			"-overlay", path,
		)
	}

	cfg.Env = append(cfg.Env,
		"GOPACKAGESDRIVER=off",
	)

	resp, err := Load(cfg, patterns)
	if err != nil {
		return nil, err
	}

	for _, pkg := range resp.Packages {
		for i, file := range pkg.GoFiles {
			path, found := overlay.ReadLink(file)
			if found {
				pkg.GoFiles[i] = path
			}
		}

		for i, file := range pkg.CompiledGoFiles {
			path, found := overlay.ReadLink(file)
			if found {
				pkg.CompiledGoFiles[i] = path
			}
		}
	}

	return resp, nil
}

func (d *Driver) ReadCacheDir(overlay fsutil.Overlay) error {
	mod, err := ReadModFile(d.workspaceDir)
	if err != nil {
		return err
	}

	dir, err := LookCacheDir(d.fsys, d.bazel, d.workspaceDir)
	if err != nil {
		return err
	}

	return WalkCacheDir(dir, func(abs, rel string) error {
		if !strings.HasSuffix(rel, ".go") {
			return nil
		}

		rel, found := CutProtoPrefix(rel)
		if !found {
			overlay.Link(filepath.Join(d.workspaceDir, rel), abs)
			return nil
		}

		if mod.Module != nil {
			rel, found := strings.CutPrefix(rel, mod.Module.Mod.Path)
			if found {
				overlay.Link(filepath.Join(d.workspaceDir, rel), abs)
				return nil
			}
		}

		return nil
	})
}
