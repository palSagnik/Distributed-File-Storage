package p2p

const (
	TypeMessage = 0x0
	TypeStream  = 0x1
)

// Message holds any arbitrary data being sent over each transport
// between two nodes in the network
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
