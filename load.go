package main

import (
	"runtime"

	"golang.org/x/tools/go/packages"
)

func Load(cfg *packages.Config, patterns []string) (*packages.DriverResponse, error) {
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}

	resp := packages.DriverResponse{
		Compiler: runtime.Compiler,
		Arch:     runtime.GOARCH,
	}

	resp.Roots = make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		resp.Roots = append(resp.Roots, pkg.ID)
	}

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		resp.Packages = append(resp.Packages, pkg)
	})

	return &resp, nil
}
