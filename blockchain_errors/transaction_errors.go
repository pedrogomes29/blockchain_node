package blockchain_errors

type ErrInvalidAddress struct{}

func (m *ErrInvalidAddress) Error() string {
	return "invalid Bitcoin Address"
}

type ErrInvalidTxInputSignature struct{}

func (m *ErrInvalidTxInputSignature) Error() string {
	return "transaction inputs have at least one invalid signature"
}

type ErrInvalidInputUTXO struct{}

func (m *ErrInvalidInputUTXO) Error() string {
	return "invalid transaction, spending from already used transaction output"
}

type ErrOutputValLGTInputVal struct{}

func (m *ErrOutputValLGTInputVal) Error() string {
	return "invalid transaction, total output value is larger than total input value"
}
