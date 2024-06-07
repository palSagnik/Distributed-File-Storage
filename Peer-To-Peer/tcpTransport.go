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

type TCPTransportConfig struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

// Transmission Control Protocol
type TCPTransport struct {
	TCPTransportConfig
	listener net.Listener

	//mutex sync.RWMutex
	//peers map[net.Addr]Peer
}

func NewTCPTransport(config TCPTransportConfig) *TCPTransport {

	return &TCPTransport{
		TCPTransportConfig: config,
	}
}

func (t *TCPTransport) ListenAndAccept() error {

	listener, err := net.Listen("tcp", t.ListenAddress)
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

	fmt.Printf("New incoming connection %+v\n", conn)
	go t.handleConnection(conn)
}

func (t *TCPTransport) handleConnection(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP Handshake Error: %s\n", err)
		return
	}

	//Reading the connection loop
	msg := &Message{}
	for {
		if err := t.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP Decoding Error: %s\n", err)
			continue
		}

		msg.From = conn.RemoteAddr()
		fmt.Printf("message: %+v\n", msg)
	}
}
