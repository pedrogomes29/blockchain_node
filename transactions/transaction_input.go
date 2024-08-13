package transactions

type TXInput struct {
	Txid      []byte
	OutIndex  int
	Signature []byte
	PubKey    []byte
}
