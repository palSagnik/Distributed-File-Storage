package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	
	peekBuffer := make([]byte, 1)
	if _, err := r.Read(peekBuffer); err != nil {
		return err
	}
	
	// If stream is true, we do not want to decode
	stream := peekBuffer[0] == TypeStream
	if stream {
		msg.Stream = true
		return nil
	}

	
	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	msg.Payload = buf[:n]

	return nil
}
