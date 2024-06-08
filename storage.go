package main

import (
	"io"
	"log"
	"os"
)

type PathTransformFunc func(string) string

var DefaultPathTransformFunc = func(key string) string {
	return key
}

type StorageConfig struct {
	PathTransformation PathTransformFunc
}

type Storage struct {
	StorageConfig
}

func NewStorage(config StorageConfig) *Storage {
	return &Storage{
		StorageConfig: config,
	}
}

func (s *Storage) writeStream(key string, r io.Reader) error {

	pathName := s.PathTransformation(key)
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	filename := "somefilename"
	completePath := pathName + "/" + filename

	f, err := os.Create(completePath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("Written %d bytes to disk: %s", n, completePath)
	return nil
}
