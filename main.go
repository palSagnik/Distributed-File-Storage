package main

import (
	"bytes"
	"log"
	"time"

	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)
func makeNewServer(listenAddress string, nodes ...string) *FileServer {

	// TCP Configuration
	tcpConfig := p2p.TCPTransportConfig{
		ListenAddress: listenAddress,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	}
	
	transport := p2p.NewTCPTransport(tcpConfig)

	fsConfig := FileServerConfig {
		StorageRoot: 		listenAddress + "-Storage",
		PathTransformation: CASPathTransformFunc,
		Transport: 			transport,
		NodeList: 			nodes,
	}

	fs := NewFileServer(fsConfig)
	transport.PeerStatus = fs.PeerStatus

	return fs
}

func main() {

	fs1 := makeNewServer(":9000", "")
	fs2 := makeNewServer(":9001", ":9000")

	go func() {
		log.Fatal(fs1.Start())
	}()
	
	time.Sleep(1 * time.Second)
	go fs2.Start()
	time.Sleep(1 * time.Second)

	key := "thisisatest"
	data := bytes.NewReader([]byte("hello this is me"))

	fs2.StoreData(key, data)
	
	select {}
}
