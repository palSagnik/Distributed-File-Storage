package main

import p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"


type FileServerConfig struct {
	StorageRoot 			string
	PathTransformation  	PathTransformFunc
	Transport 				p2p.Transport
}

type FileServer struct {
	FileServerConfig
	storage *Storage
}

func NewFileServer(config FileServerConfig) *FileServer {
	storageConfig := StorageConfig {
		Root: config.StorageRoot,
		PathTransformation: config.PathTransformation,
	}

	return &FileServer {
		FileServerConfig: config,
		storage: NewStorage(storageConfig),
	}
}

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	return nil
}