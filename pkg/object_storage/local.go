package object_storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BasePath string
}

func (s *LocalStorage) Save(_ context.Context, path string, data []byte) (string, error) {
	fullPath := filepath.Join(s.BasePath, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, data, 0644); err != nil { //nolint:gosec
		return "", err
	}
	return fullPath, nil
}

func (s *LocalStorage) Get(_ context.Context, path string) (string, error) {
	fullPath := filepath.Join(s.BasePath, path)
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("file not found")
		}
		return "", err
	}
	return fullPath, nil
}

func (s *LocalStorage) Delete(_ context.Context, path string) error {
	fullPath := filepath.Join(s.BasePath, path)
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func (s *LocalStorage) Exists(_ context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.BasePath, path)
	_, err := os.Stat(fullPath)
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
