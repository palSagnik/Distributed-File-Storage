package main

import (
	//"bytes"
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	enc "github.com/palSagnik/Distributed-File-Storage/Encoding"
	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

func makeNewServer(listenAddress string, nodes ...string) *FileServer {

	// TCP Configuration
	tcpConfig := p2p.TCPTransportConfig{
		ListenAddress: listenAddress,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	transport := p2p.NewTCPTransport(tcpConfig)

	fsConfig := FileServerConfig{
		EncryptionKey:      enc.NewEncryptionKey(),
		StorageRoot:        listenAddress + "-Storage",
		PathTransformation: CASPathTransformFunc,
		Transport:          transport,
		NodeList:           nodes,
	}

	fs := NewFileServer(fsConfig)
	transport.PeerStatus = fs.PeerStatus

	return fs
}

func main() {

	fs1 := makeNewServer(":9000", "")
	fs2 := makeNewServer(":9001", "")
	fs3 := makeNewServer(":9002", ":9001", ":9000")

	go func() { log.Fatal(fs1.Start()) }()	
	time.Sleep(2 * time.Second)
	go func() {	log.Fatal(fs2.Start()) }()

	time.Sleep(2 * time.Second)
	go fs3.Start()
	time.Sleep(2 * time.Second)

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("thisisatest_%d", i)
		data := bytes.NewReader([]byte("hello this is me"))
		fs3.Store(key, data)

		if err := fs3.storage.Delete(key); err != nil {
			log.Fatal(err)
		}
		r, err := fs3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Data: %s\n", string(b))
	}

	
}
