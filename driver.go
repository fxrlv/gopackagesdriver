package main

import (
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

type DriverConfig struct {
	Workspace string
	Bazel     string

	Workdir string
}

type Driver struct {
	workdir string
	bazel   *Bazel
}

func NewDriver(config *DriverConfig) (*Driver, error) {
	if config == nil || config.Workspace == "" {
		return &Driver{}, nil
	}

	cmd := config.Bazel
	if cmd == "" {
		cmd = "bazel"
	}

	bazel, err := NewBazel(cmd, config.Workspace)
	if err != nil {
		return nil, err
	}

	return &Driver{
		workdir: config.Workdir,
		bazel:   bazel,
	}, nil
}

func (d *Driver) Serve(req *packages.DriverRequest, patterns []string) (*packages.DriverResponse, error) {
	if d.bazel == nil {
		return &packages.DriverResponse{
			NotHandled: true,
		}, nil
	}

	rel, err := filepath.Rel(d.bazel.Workspace(), d.workdir)
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
		Dir:        d.workdir,
	}

	return d.LoadWorkspace(&cfg, patterns)
}

func (d *Driver) LoadWorkspace(cfg *packages.Config, patterns []string) (*packages.DriverResponse, error) {
	overlay, err := d.CreateOverlay(cfg.Overlay)
	if err != nil {
		return nil, err
	}

	path, cleanup, err := WriteOverlay(overlay)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	if path != "" {
		cfg.BuildFlags = append(cfg.BuildFlags,
			"-overlay", path,
		)
		cfg.Overlay = nil
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

func (d *Driver) CreateOverlay(initial map[string][]byte) (Overlay, error) {
	overlay := NewOverlay(initial)

	err := d.ReadCacheDir(overlay)
	if err != nil {
		return Overlay{}, err
	}

	return overlay, nil
}

func (d *Driver) ReadCacheDir(overlay Overlay) error {
	mod, err := ReadModFile(d.bazel.Workspace())
	if err != nil {
		return err
	}

	return d.bazel.WalkCacheDir(func(path, base string) error {
		if !strings.HasSuffix(base, ".go") {
			return nil
		}

		base, found := CutProtoPrefix(base)
		if !found {
			overlay.Link(filepath.Join(d.bazel.Workspace(), base), path)
			return nil
		}

		if mod.Module != nil {
			base, found := strings.CutPrefix(base, mod.Module.Mod.Path)
			if found {
				overlay.Link(filepath.Join(d.bazel.Workspace(), base), path)
				return nil
			}
		}

		return nil
	})
}
