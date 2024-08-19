package server

type commandID int

const (
	GET_ADDR commandID = iota
	ADDR
	VERSION
	VERSION_ACK
	GET_INV
	INV
	GET_DATA
	BLOCK
	TX
)

type command struct {
	id     commandID
	peer   *peer
	args   [][]byte
}