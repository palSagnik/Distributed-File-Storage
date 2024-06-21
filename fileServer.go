package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	enc "github.com/palSagnik/Distributed-File-Storage/Encoding"
	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

type FileServerConfig struct {
	EncryptionKey      []byte
	StorageRoot        string
	PathTransformation PathTransformFunc
	Transport          p2p.Transport
	NodeList           []string
}

type FileServer struct {
	FileServerConfig

	lockPeer sync.Mutex
	peers    map[string]p2p.Peer

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
		peers:            make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
	Key string
}

func (fs *FileServer) Get(key string) (io.Reader, error) {

	if fs.storage.Present(key) {
		fmt.Printf("[%s] File (%s) found locally.\n", fs.Transport.Addr(), key)

		_, r, err := fs.storage.Read(key)
		return r, err
	}

	fmt.Printf("[%s] No file (%s) found locally. Searching in network\n", fs.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}

	if err := fs.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range fs.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		n, err := fs.storage.WriteDecrypt(fs.EncryptionKey, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err // TODO: handle error properly
		}

		fmt.Printf("[%s] Received %d bytes over the network from (%s)\n", fs.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	_, r, err := fs.storage.Read(key)
	return r, err
}

func (fs *FileServer) Store(key string, r io.Reader) error {

	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)
	size, err := fs.storage.writeStream(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size + 16,
		},
	}

	if err := fs.broadcast(&msg); err != nil {
		return err

	}

	time.Sleep(20 * time.Millisecond)

	peers := []io.Writer{}
	for _, peer := range fs.peers {
		peers = append(peers, peer)
	}
	multiwrite := io.MultiWriter(peers...)
	multiwrite.Write([]byte{p2p.TypeStream})

	n, err := enc.StreamEncrypt(fs.EncryptionKey, fileBuffer, multiwrite)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Received and written (%d) bytes to disk\n", fs.Transport.Addr(), n)

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
		case rpc := <-fs.Transport.Consume():
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m); err != nil {
				log.Printf("Decoding Error: %s", err)
			}

			if err := fs.handleMessage(rpc.From, &m); err != nil {
				log.Printf("Handle Message Error: %s", err)
			}
		case <-fs.quitChannel:
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

func (fs *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !fs.storage.Present(msg.Key) {
		return fmt.Errorf("[%s] no such file (%s) found on local disk", fs.Transport.Addr(), msg.Key)
	}

	fmt.Printf("[%s] File (%s) found. Sending over the network\n", fs.Transport.Addr(), msg.Key)
	fileSize, r, err := fs.storage.Read(msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		defer rc.Close()
	}

	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found in map", from)
	}

	// First byte -> Type of stream
	// Second byte -> Size of stream
	peer.Send([]byte{p2p.TypeStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written (%d) over the network to %s\n", fs.Transport.Addr(), n, from)
	return nil
}

func (fs *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {

	peer, ok := fs.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found", from)
	}

	n, err := fs.storage.writeStream(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}
	fmt.Printf("[%s] Written %d bytes to disk\n", fs.Transport.Addr(), n)
	peer.CloseStream()

	return nil
}

func (fs *FileServer) listedNodeNetwork() error {

	for _, addr := range fs.NodeList {

		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			log.Printf("[%s] Attempting to connect %s\n", fs.Transport.Addr(), addr)
			if err := fs.Transport.Dial(addr); err != nil {
				log.Println("Dial Error ", err)
			} else {
				log.Printf("Connected to %s", addr)
			}
		}(addr)
		time.Sleep(time.Second)
	}

	return nil
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
	return gob.NewEncoder(multiwrite).Encode(msg)
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
