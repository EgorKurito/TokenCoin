package blockchain

type TxOutput struct {
	Value  int
	PubKey string
}

func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
