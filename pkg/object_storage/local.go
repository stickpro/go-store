package object_storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	BasePath string
}

// resolve turns a stored/relative path into a real filesystem path. It accepts
// both a path relative to BasePath ("public/images/x.jpg") and one that already
// includes it ("storage/public/images/x.jpg", as historically stored in media.path).
func (s *LocalStorage) resolve(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	clean := filepath.Clean(path)
	base := filepath.Clean(s.BasePath)
	if clean == base || strings.HasPrefix(clean, base+string(filepath.Separator)) {
		return clean
	}
	return filepath.Join(base, clean)
}

func (s *LocalStorage) Save(_ context.Context, path string, data []byte) (string, error) {
	fullPath := s.resolve(path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil { //nolint:gosec
		return "", err
	}
	return fullPath, nil
}

func (s *LocalStorage) Get(_ context.Context, path string) (string, error) {
	fullPath := s.resolve(path)
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("file not found")
		}
		return "", err
	}
	return fullPath, nil
}

func (s *LocalStorage) Read(_ context.Context, path string) ([]byte, error) {
	data, err := os.ReadFile(s.resolve(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return data, nil
}

func (s *LocalStorage) Delete(_ context.Context, path string) error {
	if err := os.Remove(s.resolve(path)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func (s *LocalStorage) Exists(_ context.Context, path string) (bool, error) {
	_, err := os.Stat(s.resolve(path))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *LocalStorage) URL(_ context.Context, path string) (string, error) {
	return path, nil
}

func New(basePath string) *LocalStorage {
	return &LocalStorage{
		BasePath: basePath,
	}
}
