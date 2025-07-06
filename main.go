package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"golang.org/x/tools/go/packages"
)

var (
	initOverlay = flag.Bool("init", false, "init overlay")
	syncOverlay = flag.Bool("sync", false, "sync overlay")
)

func main() {
	flag.Parse()

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

	if *initOverlay || *syncOverlay {
		path, err := driver.MakeOverlay(*syncOverlay)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(path)
		return
	}

	err = serve(driver, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serve(d *Driver, args []string) error {
	var req packages.DriverRequest

	err := json.NewDecoder(os.Stdin).Decode(&req)
	if err != nil {
		return err
	}

	resp, err := d.Serve(&req, args)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(resp)
}
