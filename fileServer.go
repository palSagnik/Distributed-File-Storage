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
	Payload any
}

type MessageStoreFile struct {
	Key 	string
	Size 	int64
}

func (fs *FileServer) StoreData(key string, r io.Reader) error {

	buffer := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
			Size: 16,
		},
	}

	if err := gob.NewEncoder(buffer).Encode(msg); err != nil {
		return err
	}

	for _, peers := range fs.peers {
		if err := peers.Send(buffer.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(3 * time.Second)

	//payload := []byte("THIS IS A LARGE FILE")
	for _, peer := range fs.peers {
		_, err := io.Copy(peer, r)
		if err != nil {
			return err
		}
	}

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
		case rpc:= <- fs.Transport.Consume():
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m); err != nil {
				log.Fatal(err)
			}
			
			if err := fs.handleMessage(rpc.From, &m); err != nil {
				log.Println(err)
				return
			}
		case <- fs.quitChannel:
			return
		}
	}
}

func (fs *FileServer) handleMessage(from string, m *Message) error {
	switch payload := m.Payload.(type) {
	case MessageStoreFile:
		return fs.handleMessageStoreFile(from, payload)
	}
	return nil
}

func (fs *FileServer) handleMessageStoreFile (from string, msg MessageStoreFile) error {
	
	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found", from)
	}

	if err := fs.storage.writeStream(msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
		return err
	}
	
	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
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

// func (fs *FileServer) broadcast(msg *Message) error {
	
// 	peers := []io.Writer{}
// 	for _, peer := range fs.peers {
// 			peers = append(peers, peer)
// 		}
// 	multiwrite := io.MultiWriter(peers...)
// 	return gob.NewEncoder(multiwrite).Encode(msg.Payload)
// }

func init() {
	gob.Register(MessageStoreFile{})
}