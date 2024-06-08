package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTCPTransport(t *testing.T) {

	tcpConfig := TCPTransportConfig{
		ListenAddress: ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}

	transport := NewTCPTransport(tcpConfig)

	assert.Equal(t, transport.ListenAddress, ":3000")

	//Server
	assert.Nil(t, transport.ListenAndAccept())
}
