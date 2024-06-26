package p2p

import "net"

// Peer is a remote node
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Tranport handles the communication between nodes.
// Or in simple words they can be of form like
// TCP, UDP, websockets ...
type Transport interface {
	Addr()	string
	Dial(address string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
