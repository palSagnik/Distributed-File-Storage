package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {

	tcpConfig := TCPTransportConfig{
		ListenAddress: ":9000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}

	transport := NewTCPTransport(tcpConfig)

	assert.Equal(t, transport.ListenAddress, ":9000")

	//Server
	assert.Nil(t, transport.ListenAndAccept())
}
