package main

import (
	"log"
	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

func main() {
	
	// TCP Configuration
	tcpConfig := p2p.TCPTransportConfig{
		ListenAddress: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
		// Need On Peer
	}
	
	transport := p2p.NewTCPTransport(tcpConfig)

	fsConfig := FileServerConfig {
		StorageRoot: ":3000-Storage",
		PathTransformation: CASPathTransformFunc,
		Transport: transport,
	}

	fs := NewFileServer(fsConfig)
	if err := fs.Start(); err != nil {
		log.Fatal(err)
	}

	select {}
}
