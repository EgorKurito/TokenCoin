package blockchain

// OutPoint is token data
type OutPoint struct {
	ID   int
	Hash []byte
}

// NewOutPoint return a new token transaction
func NewOutPoint(hash *[]byte, index int) *OutPoint {
	return &OutPoint{
		ID:   index,
		Hash: *hash,
	}
}

// TxInput represents a transaction input
type TxInput struct {
	PreviousOutPoint OutPoint
	PubKey           []byte
	Signature        []byte
}

// NewTxInput create a new TxInput
func NewTxInput(prevOut *OutPoint, sign, pubKey []byte) *TxInput {
	newTxInput := &TxInput{
		PreviousOutPoint: *prevOut,
		PubKey:           pubKey,
		Signature:        sign,
	}

	return newTxInput
}
