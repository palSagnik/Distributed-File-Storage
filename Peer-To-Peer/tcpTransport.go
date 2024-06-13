package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
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

func (p *TCPPeer) Send(data []byte) error {
	_, err := p.connection.Write(data)
	return err
}

func (p *TCPPeer) RemoteAddr() net.Addr {
	return p.connection.RemoteAddr()
}

// Close implements the closure of the peer interface
func (p *TCPPeer) Close() error {
	return p.connection.Close()
}

// This struct type defines the configuration for a particular transport.
// ListenAddress -> The address of the transport to connect to.
// HandSHakeFunc -> For initiating Handshake as we see in a TCP model.
// Decoder -> The type of decoder to use based on the transport.
// PeerStatus -> If the func returns an error we drop the peer.
type TCPTransportConfig struct {
	ListenAddress string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	PeerStatus    func(Peer) error
}

// Transmission Control Protocol
type TCPTransport struct {
	TCPTransportConfig
	listener   net.Listener
	rpcChannel chan RPC
}

func NewTCPTransport(config TCPTransportConfig) *TCPTransport {
	return &TCPTransport{
		TCPTransportConfig: config,
		rpcChannel:         make(chan RPC),
	}
}

// Consume implements transport interface which will only return
// a read-only channel for incoming messages from other peers
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChannel
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Dial(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	go t.handleConnection(conn, true) // because we are dialing: outbound -> true
	return nil
}


func (t *TCPTransport) ListenAndAccept() error {

	listener, err := net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	//Accept part
	t.listener = listener
	go t.acceptLoop()

	log.Printf("TCP transport listening on: %s\n", t.ListenAddress)
	return nil
}

// private func for accept loop
func (t *TCPTransport) acceptLoop() {

	// is for loop necessary?
	conn, err := t.listener.Accept()

	// If the connection is closed, we return
	if errors.Is(err, net.ErrClosed) {
		return
	}
	if err != nil {
		fmt.Println("TCP Connection Error")
	}

	go t.handleConnection(conn, false) // because we are accepting the connection
}

func (t *TCPTransport) handleConnection(conn net.Conn, outbound bool) {
	var err error

	// after generating any kind of error the handleConnection function exits
	// then defer is called which drops the peer connection
	// this is true for any error
	defer func() {
		fmt.Printf("dropping peer connection: %+v", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)
	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP Handshake Error: %s\n", err)
		return
	}

	// If someone provides the PeerStatus Func
	// And it doesnt give an error, then we move to the read loop
	if t.PeerStatus != nil {
		if err = t.PeerStatus(peer); err != nil {
			return
		}
	}

	//Reading the connection loop after the handshake and peerStatus doesnt fail
	rpc := RPC{}
	for {
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			fmt.Printf("TCP Decoding Error: %s\n", err)
			continue
		}

		rpc.From = conn.RemoteAddr()
		t.rpcChannel <- rpc
	}
}
