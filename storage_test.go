package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
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
	key := "mygreatestgoal"
	data := []byte("this will work")

	// Create Operation
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}


	// Present Operation
	if ok := s.Present(key); !ok {
		t.Errorf("expected to have key %s, but not found", key)
	}	


	// Read Operation
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("wanted %s, got %s", data, b)
	}

	fmt.Println(string(b))


	// Delete Operation
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}
