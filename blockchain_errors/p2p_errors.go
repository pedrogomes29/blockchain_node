package blockchain_errors

type ErrOrphanBlock struct{}

func (m *ErrOrphanBlock) Error() string {
	return "received block is an orphan"
}
