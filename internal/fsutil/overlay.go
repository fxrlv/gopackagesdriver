package fsutil

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
)

type Overlay struct {
	replace map[string]string
}

func NewOverlay() Overlay {
	return Overlay{
		replace: make(map[string]string),
	}
}

func (o Overlay) Len() int {
	return len(o.replace)
}

func (o Overlay) Link(target, source string) {
	o.replace[target] = source
}

func (o Overlay) ReadLink(name string) (string, bool) {
	source, found := o.replace[name]
	return source, found
}

func (o Overlay) Append(dir string, content map[string][]byte) (Overlay, error) {
	o.replace = maps.Clone(o.replace)

	for target, data := range content {
		source := filepath.Join(dir, fmt.Sprintf(
			"%d-%s", o.Len(), filepath.Base(target),
		))

		err := os.WriteFile(source, data, 0600)
		if err != nil {
			return Overlay{}, err
		}

		o.replace[target] = source
	}

	return o, nil
}

type overlayJSON struct {
	Replace map[string]string `json:"replace"`
}

func MarshalOverlay(o Overlay) ([]byte, error) {
	return json.Marshal(overlayJSON{o.replace})
}

func UnmarshalOverlay(data []byte) (Overlay, error) {
	var v overlayJSON
	err := json.Unmarshal(data, &v)
	return Overlay{v.Replace}, err
}
