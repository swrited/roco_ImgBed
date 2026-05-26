package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalAdapter struct {
	root string
	url  string
}

func NewLocalAdapter(root, url string) *LocalAdapter {
	if root == "" {
		root = "uploads"
	}
	return &LocalAdapter{root: root, url: url}
}

func (a *LocalAdapter) Save(path string, data []byte) error {
	fullPath := filepath.Join(a.root, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (a *LocalAdapter) Open(path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(a.root, path))
}

func (a *LocalAdapter) SetPublic(_ string, _ bool) error {
	return nil
}

func (a *LocalAdapter) Delete(path string) error {
	fullPath := filepath.Join(a.root, path)
	return os.Remove(fullPath)
}

func (a *LocalAdapter) Exists(path string) bool {
	_, err := os.Stat(filepath.Join(a.root, path))
	return err == nil
}

func (a *LocalAdapter) URL(path string) string {
	url := a.url
	if url == "" {
		url = "/"
	}
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url + path
}
