package p2p

import (
	"fmt"
	"net"
	//"sync"
)

// TCPPeer represents the remote node in established TCP connection
type TCPPeer struct {

	// This is the connection of the Peer
	connection net.Conn

	// if we dial and retrieve a connection => outbound == true
	// if we accept and retrieve a connection => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {

	return &TCPPeer{
		connection: conn,
		outbound:   outbound,
	}
}

// Transmission Control Protocol
type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	handshakeFunc HandshakeFunc
	decoder       Decoder

	//mutex sync.RWMutex
	//peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {

	return &TCPTransport{
		handshakeFunc: NOPHandshakeFunc,
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {

	listener, err := net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	//Accept part
	t.listener = listener
	go t.acceptLoop()

	return nil
}

// private func for accept loop
func (t *TCPTransport) acceptLoop() {
	conn, err := t.listener.Accept()
	if err != nil {
		fmt.Println("TCP Connection Error")
	}
	go t.handleConnection(conn)
}

type Temp struct{}

func (t *TCPTransport) handleConnection(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.handshakeFunc(peer); err != nil {

	}

	//Reading the connection loop
	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP Decoding Error: %s\n", err)
			continue
		}

	}
}
