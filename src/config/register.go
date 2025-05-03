package config

import (
	"encoding/json"
	"os"
	"sync"
)

type FileRegisterEntry struct {
	Path      string `json:"path"`
	Private   bool   `json:"private"`
	UUID      string `json:"uuid"`
	Owner     string `json:"owner"`
	Timestamp string `json:"timestamp"`
}

type FileRegister struct {
	mu    sync.Mutex
	Items map[string]FileRegisterEntry
}

func LoadRegister(path string) (*FileRegister, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// create file
			err := os.WriteFile(path, []byte("{}"), 0644)
			if err != nil {
				return nil, err
			}

			return &FileRegister{Items: make(map[string]FileRegisterEntry)}, nil
		}
		return nil, err
	}

	var items map[string]FileRegisterEntry
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return &FileRegister{Items: items}, nil
}

func (r *FileRegister) Save(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := json.MarshalIndent(r.Items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (r *FileRegister) Add(name string, entry FileRegisterEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Items[name] = entry
}

func (r *FileRegister) Get(name string) (FileRegisterEntry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.Items[name]
	return entry, ok
}
