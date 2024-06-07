package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTCPTransport(t *testing.T) {

	tcpConfig := p2p.TCPTransportConfig{
		ListenAddress: ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
	}

	transport := NewTCPTransport(tcpConfig)

	assert.Equal(t, transport.ListenAddress, listenAddr)

	//Server
	assert.Nil(t, transport.ListenAndAccept())
}
