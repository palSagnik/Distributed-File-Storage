package p2p

// Peer is a remote node
type Peer interface {
	Close() error
}

// Tranport handles the communication between nodes.
// Or in simple words they can be of form like
// TCP, UDP, websockets ...
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
