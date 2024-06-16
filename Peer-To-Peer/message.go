package p2p


// Message holds any arbitrary data being sent over each transport
// between two nodes in the network
type RPC struct {
	From    string
	Payload []byte
}
