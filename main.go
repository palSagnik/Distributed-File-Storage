package main

import (
	"log"

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
	
	fs2.Start()

}
