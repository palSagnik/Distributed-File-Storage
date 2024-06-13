package main

import (
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

type FileServerConfig struct {
	StorageRoot        string
	PathTransformation PathTransformFunc
	Transport          p2p.Transport
	NodeList 		   []string
}

type FileServer struct {
	FileServerConfig

	lockPeer 		sync.Mutex
	peers 			map[string]p2p.Peer

	storage     	*Storage
	quitChannel 	chan struct{}
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
		peers:			  make(map[string]p2p.Peer),
	}
}

type Payload struct {
	Key 	string
	Data 	[]byte
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {


	return nil
}


func (fs *FileServer) PeerStatus(p p2p.Peer) error {
	fs.lockPeer.Lock()
	defer fs.lockPeer.Unlock()

	fs.peers[p.RemoteAddr().String()] = p
	log.Printf("Connected with remote %s", p.RemoteAddr())

	return nil
}

func (fs *FileServer) Start() error {
	if err := fs.Transport.ListenAndAccept(); err != nil {
		return err
	}

	fs.listedNodeNetwork()
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

func (fs *FileServer) listedNodeNetwork() error {

	for _, addr := range fs.NodeList {

		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			log.Println("Attempting to connect", addr)
			if err := fs.Transport.Dial(addr); err != nil {
				log.Println("Dial Error ",  err)
			} else {
				log.Printf("Connected to %s", addr)
			}
		} (addr)
		time.Sleep(time.Second)
	}

	return nil;
}