package main

import (
	"os"

	"github.com/fxrlv/gopackagesdriver/internal/fsutil"
)

func ReadOverlay(path string) (fsutil.Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return fsutil.Overlay{}, err
	}

	return fsutil.UnmarshalOverlay(data)
}

func WriteOverlay(dir string, overlay fsutil.Overlay) (string, error) {
	data, err := fsutil.MarshalOverlay(overlay)
	if err != nil {
		return "", err
	}

	f, err := os.CreateTemp(dir, "*-overlay.json")
	if err != nil {
		return "", err
	}

	_, err = f.Write(data)
	if err != nil {
		f.Close()
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}
