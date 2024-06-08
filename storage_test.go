package main

import (
	"bytes"
	"testing"
)

func TestStorage(t *testing.T) {
	config := StorageConfig{
		PathTransformation: DefaultPathTransformFunc,
	}

	s := NewStorage(config)
	key := "mygotesthere"
	data := bytes.NewReader([]byte("this will work"))

	if err := s.writeStream(key, data); err != nil {
		t.Error(err)
	}

}
