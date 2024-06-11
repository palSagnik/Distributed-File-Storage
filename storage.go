package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

//defining the default root folder
const defaultRootFolder = "networkStorage"

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashString := hex.EncodeToString(hash[:])

	blockSize := 8
	sliceLength := len(hashString) / blockSize

	// Goes over the hashString and breaks it into a slice of substrings
	// This is then used as path for the file
	// blockSize -> size of each substring
	// sliceLength -> number of substrings
	paths := make([]string, sliceLength)
	for i := 0; i < sliceLength; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashString[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashString,
		PathRoot: 	  paths[0],
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
	PathRoot     string
}

func (p PathKey) CompletePath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type StorageConfig struct {

	//Root is folder name of the root containing the files and folder on the disk
	Root 			   string
	PathTransformation PathTransformFunc
}

type Storage struct {
	StorageConfig
}

func NewStorage(config StorageConfig) *Storage {
	
	// assigning default values
	if config.PathTransformation == nil {
		config.PathTransformation = DefaultPathTransformFunc
	}

	if config.Root == "" {
		config.Root = defaultRootFolder
	}

	return &Storage{
		StorageConfig: config,
	}
}

func (s *Storage) Present(key string) bool {
	pk := s.PathTransformation(key)
	fullpath := fmt.Sprintf("%s/%s", s.Root, pk.CompletePath())

	_, err := os.Stat(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			fmt.Printf("Error in statting the file: %s", err)
		}
	}

	return true
}

// This delete function for now is not taking in account
// the probabibility of partial hash collision
func (s *Storage) Delete(key string) error {
	pk := s.PathTransformation(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pk.Filename)
	}()

	completePath := fmt.Sprintf("%s/%s", s.Root, pk.PathRoot)
	if err := os.RemoveAll(completePath); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, f)

	return buffer, err
}

func (s *Storage) readStream(key string) (io.ReadCloser, error) {
	pk := s.PathTransformation(key)

	completePath := fmt.Sprintf("%s/%s", s.Root, pk.CompletePath())
	return os.Open(completePath)

}

func (s *Storage) writeStream(key string, r io.Reader) error {

	pathKey := s.PathTransformation(key)
	pathName := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}

	completePath := fmt.Sprintf("%s/%s", s.Root, pathKey.CompletePath())
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
