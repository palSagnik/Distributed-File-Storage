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

type MessageGetFile struct {
	Key string
}

func (fs *FileServer) Get(key string) (io.Reader, error) {

	if fs.storage.Present(key) {
		return fs.storage.Read(key)
	}

	fmt.Printf("No file (%s) found locally. Searching in network\n", key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := fs.broadcast(&msg); err != nil {
		return nil, err
	}

	for _, peer := range fs.peers {
		fileBuffer := new(bytes.Buffer)
		n, err := io.Copy(fileBuffer, peer)
		if err != nil {
			return nil, err		// TODO: handle error properly
		}

		fmt.Printf("Received %d bytes over the network\n", n)
		fmt.Println(fileBuffer.String())
	}
	select {}
	return nil, nil
}

func (fs *FileServer) Store(key string, r io.Reader) error {

	var (
		fileBuffer = new(bytes.Buffer)
		tee = io.TeeReader(r, fileBuffer)
	)
	size, err := fs.storage.writeStream(key, tee)
	if err != nil {
		return err 
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
			Size: size,
		},
	}

	if err := fs.broadcast(&msg); err != nil {
		return err

	}

	time.Sleep(20 * time.Millisecond)

	for _, peer := range fs.peers {
		peer.Send([]byte{p2p.TypeStream})
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}

		fmt.Printf("Received and written %d bytes\n", n)
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
				log.Printf("Decoding Error: ", err)
			}
			
			if err := fs.handleMessage(rpc.From, &m); err != nil {
				log.Printf("Handle Message Error: ", err)
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
	case MessageGetFile:
		return fs.handleMessageGetFile(from, payload)
	}
	return nil
}

func (fs *FileServer) handleMessageGetFile (from string, msg MessageGetFile) error {
	fmt.Println("Fetching file from network")

	if !fs.storage.Present(msg.Key) {
		return fmt.Errorf("no such key (%s) found", msg.Key)
	}
	r, err := fs.storage.Read(msg.Key)
	if err != nil {
		return err
	}

	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found in map", from)
	}

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("Read %d over the network from %s\n", n, from)
	return nil
}
func (fs *FileServer) handleMessageStoreFile (from string, msg MessageStoreFile) error {
	
	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found", from)
	}

	n, err := fs.storage.writeStream(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	
	log.Printf("Written %d bytes to disk\n", n)
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


func (fs *FileServer) broadcast(msg *Message) error {

	msgBuffer := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuffer).Encode(msg); err != nil {
		return err
	}

	for _, peer := range fs.peers {
		peer.Send([]byte{p2p.TypeMessage})
		if err := peer.Send(msgBuffer.Bytes()); err != nil {
			return err
		}
	}

	return nil
}


func (fs *FileServer) stream(msg *Message) error {
	
	peers := []io.Writer{}
	for _, peer := range fs.peers {
			peer.Send([]byte{p2p.TypeStream})
			peers = append(peers, peer)
		}
	multiwrite := io.MultiWriter(peers...)
	return gob.NewEncoder(multiwrite).Encode(msg.Payload)
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}