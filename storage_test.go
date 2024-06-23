package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	enc "github.com/palSagnik/Distributed-File-Storage/Encoding"
)

func TestPathTransformFunc(t *testing.T) {
	key := "mygreatestgoal"
	pathkey := CASPathTransformFunc(key)
	expectedFilenamekey := "bef06e07b129dc25191b1ccffffb0f6d05e64125"
	expectedPathName := "bef06e07/b129dc25/191b1ccf/fffb0f6d/05e64125"

	if pathkey.PathName != expectedPathName {
		t.Errorf("wanted %s, got %s", expectedPathName, pathkey.PathName)
	}
	if pathkey.Filename != expectedFilenamekey {
		t.Errorf("wanted %s, got %s", expectedFilenamekey, pathkey.Filename)
	}
	fmt.Println(pathkey.PathName)
}

func TestStorageCRD(t *testing.T) {
	config := StorageConfig{
		PathTransformation: CASPathTransformFunc,
	}
	s := NewStorage(config)
	id := enc.GenerateID()
	defer teardown(t, s)

	key := "fooandbar"
	data := []byte("this will work bruh")

	// Create
	if _, err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	// Present
	if ok := s.Present(id, key); !ok {
		t.Errorf("expected to have key %s, but not found", key)
	}	


	// Read
	_, r, err := s.Read(id, key)
	if err != nil {
		t.Error(err)
	}
	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("wanted %s, got %s", data, b)
	}

	fmt.Println(string(b))


	// Delete
	if err := s.Delete(id, key); err != nil {
		t.Error(err)
	}
}

func teardown(t *testing.T, s *Storage) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}