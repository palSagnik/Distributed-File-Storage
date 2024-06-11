package main

import (
	"fmt"

	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

type FileServerConfig struct {
	StorageRoot        string
	PathTransformation PathTransformFunc
	Transport          p2p.Transport
}

type FileServer struct {
	FileServerConfig
	storage     *Storage
	quitChannel chan struct{}
}

func NewFileServer(config FileServerConfig) *FileServer {
	storageConfig := StorageConfig{
		Root:               config.StorageRoot,
		PathTransformation: config.PathTransformation,
	}

	return &FileServer{
		FileServerConfig: config,
		storage:          NewStorage(storageConfig),
		quitChannel:      make(chan struct{}),
	}
}

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.loop()

	return nil
}

func (fs *FileServer) Stop() {
	close(fs.quitChannel)
}

func (fs *FileServer) loop() {

	defer func() {
		fmt.Println("file server stopped")
		fs.Transport.Close()
	}()

	for {
		select {
		case msg:= <- fs.Transport.Consume():
			fmt.Println(msg)
		case <- fs.quitChannel:
			return
		}
	}
}