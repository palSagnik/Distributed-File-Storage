package main

import (
	"log"

	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

func main() {

	tcpConfig := p2p.TCPTransportConfig{
		ListenAddress: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	transport := p2p.NewTCPTransport(tcpConfig)
	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
