package blockchain_errors

type ErrInvalidAddress struct {}

func (m *ErrInvalidAddress) Error() string {
	return "Invalid Bitcoin Address"
}


type ErrInvalidTxInputSignature struct {}

func (m *ErrInvalidTxInputSignature) Error() string {
	return "Transaction inputs have at least one invalid signature"
}