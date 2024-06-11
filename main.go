package main

import (
	"log"
	"time"

	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

func main() {

	// TCP Configuration
	tcpConfig := p2p.TCPTransportConfig{
		ListenAddress: ":9000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
		// Need On Peer
	}
	
	transport := p2p.NewTCPTransport(tcpConfig)

	fsConfig := FileServerConfig {
		StorageRoot: ":9000-Storage",
		PathTransformation: CASPathTransformFunc,
		Transport: transport,
	}

	fs := NewFileServer(fsConfig)

	go func(){
		time.Sleep(time.Second * 3)
		fs.Stop()
	}()

	if err := fs.Start(); err != nil {
		log.Fatal(err)
	}

}
