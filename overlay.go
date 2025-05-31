package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
)

type Overlay struct {
	overlay map[string][]byte
	replace map[string]string
}

func NewOverlay(initial map[string][]byte) Overlay {
	return Overlay{
		overlay: maps.Clone(initial),
		replace: make(map[string]string),
	}
}

func (o Overlay) Link(name, source string) {
	delete(o.overlay, name)
	o.replace[name] = source
}

func (o Overlay) ReadLink(name string) (string, bool) {
	source, found := o.replace[name]
	return source, found
}

func WriteOverlay(o Overlay) (path string, cleanup func(), err error) {
	type OverlayJSON struct {
		Replace map[string]string `json:"replace"`
	}

	if len(o.overlay) == 0 {
		if len(o.replace) == 0 {
			return "", func() {}, nil
		}

		data, err := json.Marshal(OverlayJSON{o.replace})
		if err != nil {
			return "", nil, err
		}

		file, err := WriteTempFile("gopackagesdriver.*.overlay.json", data)
		if err != nil {
			return "", nil, err
		}

		return file, func() { os.Remove(file) }, nil
	}

	dir, err := MkdirTemp("gopackagesdriver.*.overlay")
	if err != nil {
		return "", nil, err
	}

	defer func() {
		cleanup = func() {
			os.RemoveAll(dir)
		}

		if err != nil {
			cleanup()
			cleanup = nil
		}
	}()

	replace := maps.Clone(o.replace)
	for name, data := range o.overlay {
		base := fmt.Sprintf("%d-%s",
			len(replace), filepath.Base(name),
		)
		temp := filepath.Join(dir, base)

		err = os.WriteFile(temp, data, 0600)
		if err != nil {
			return "", nil, err
		}

		replace[name] = temp
	}

	data, err := json.Marshal(OverlayJSON{replace})
	if err != nil {
		return "", nil, err
	}

	file := filepath.Join(
		dir, "overlay.json",
	)

	err = os.WriteFile(file, data, 0600)
	if err != nil {
		return "", nil, err
	}

	return file, nil, nil
}
