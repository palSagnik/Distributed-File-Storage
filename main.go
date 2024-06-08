package main

import (
	"fmt"
	"log"

	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

func PeerStatus(p2p.Peer) error {
	//return fmt.Errorf("failed Peer Status func")
	return nil
}

func main() {

	tcpConfig := p2p.TCPTransportConfig{
		ListenAddress: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		PeerStatus:    PeerStatus,
	}

	transport := p2p.NewTCPTransport(tcpConfig)

	go func() {
		for {
			msg := <-transport.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
