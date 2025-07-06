package main

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	var req packages.DriverRequest

	err := json.NewDecoder(os.Stdin).Decode(&req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	workdir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	config := DriverConfig{
		WorkspaceDir: os.Getenv("GOPACKAGESDRIVER_WORKSPACE"),
		WorkingDir:   workdir,
		Bazel:        os.Getenv("GOPACKAGESDRIVER_BAZEL"),
	}

	driver, err := NewDriver(&config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	resp, err := driver.Serve(&req, os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = json.NewEncoder(os.Stdout).Encode(resp)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
