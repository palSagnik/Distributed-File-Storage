package main

import (
	"bytes"
	"encoding/gob"
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

type Message struct {
	From string
	Payload any
}

type MessageData struct {
	Key 	string
	Data 	[]byte
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {

	buffer := new(bytes.Buffer)
	tee := io.TeeReader(r, buffer)

	// Store data in the disk
	if err := fs.storage.Write(key, tee); err != nil {
		return err
	}

	p := &MessageData{
		Key: key,
		Data: buffer.Bytes(),
	}

	fmt.Printf("Broadcasting: %+v\n", buffer.String())
	return fs.broadcast(&Message{
		From: "todo",
		Payload: p,
	})
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
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&m); err != nil {
				log.Fatal(err)
			}
			
			if err := fs.handleMessage(&m); err != nil {
				log.Fatal(err)
			}
		case <- fs.quitChannel:
			return
		}
	}
}

func (fs *FileServer) handleMessage(msg *Message) error {

	
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

func (fs *FileServer) broadcast(msg *Message) error {
	
	peers := []io.Writer{}
	for _, peer := range fs.peers {
			peers = append(peers, peer)
		}
	multiwrite := io.MultiWriter(peers...)
	return gob.NewEncoder(multiwrite).Encode(msg.Payload)
}