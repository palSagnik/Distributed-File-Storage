package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	enc "github.com/palSagnik/Distributed-File-Storage/Encoding"
)

// defining the default root folder
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
		PathRoot: paths[0],
	}
}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	PathName string
	Filename string
	PathRoot string
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
	Root               string
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

func (s *Storage) Present(id string, key string) bool {
	pk := s.PathTransformation(key)
	fullpath := fmt.Sprintf("%s/%s/%s", s.Root, id, pk.CompletePath())

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

// Clearing the entire storage along with the root folder
func (s *Storage) Clear() error {
	return os.RemoveAll(s.Root)
}

// This delete function for now is not taking in account
// the probabibility of partial hash collision
func (s *Storage) Delete(id string, key string) error {
	pk := s.PathTransformation(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pk.Filename)
	}()

	completePath := fmt.Sprintf("%s/%s/%s", s.Root, id, pk.PathRoot)
	if err := os.RemoveAll(completePath); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Write(id string, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

func (s *Storage) WriteDecrypt(id string, encKey []byte, key string, r io.Reader) (int64, error) {
	f, err := s.openFile(id, key)
	if err != nil {
		return 0, err
	}

	n, err := enc.StreamDecrypt(encKey, r, f)
	return int64(n), err
}

func (s *Storage) openFile(id string, key string) (*os.File, error) {
	pathKey := s.PathTransformation(key)
	pathName := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)

	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return nil, err
	}

	completePath := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.CompletePath())
	return os.Create(completePath)

}
func (s *Storage) writeStream(id string, key string, r io.Reader) (int64, error) {

	f, err := s.openFile(id, key)
	if err != nil {
		return 0, err
	}

	return io.Copy(f, r)
}

func (s *Storage) Read(id string, key string) (int64, io.Reader, error) {
	return s.readStream(id, key)
}

func (s *Storage) readStream(id string, key string) (int64, io.ReadCloser, error) {
	pk := s.PathTransformation(key)
	completePath := fmt.Sprintf("%s/%s/%s", s.Root, id, pk.CompletePath())

	file, err := os.Open(completePath)
	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), file, nil
}
