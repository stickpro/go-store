package object_storage

import (
	"context"
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
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", err
	}
	return fullPath, nil
}

func (s *LocalStorage) Get(ctx context.Context, path string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.BasePath, path)
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *LocalStorage) URL(ctx context.Context, path string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func New(basePath string) *LocalStorage {
	return &LocalStorage{
		BasePath: basePath,
	}
}
