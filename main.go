package main

import (
	"log"

	p2p "github.com/palSagnik/Distributed-File-Storage/Peer-To-Peer"
)

func main() {

	transport := p2p.NewTCPTransport(":3000")
	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
