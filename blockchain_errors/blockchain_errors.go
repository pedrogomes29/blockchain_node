package blockchain_errors

type ErrInvalidAddress struct {}

func (m *ErrInvalidAddress) Error() string {
	return "Invalid Bitcoin Address"
}