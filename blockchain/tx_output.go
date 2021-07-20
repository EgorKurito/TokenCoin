package blockchain

import (
	"bytes"

	"github.com/EgorKurito/TokenCoin/blockchain/wallet"
)

// TxOutput represents a transaction output
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// CanBeUnlocked checked if the output can be used by the owner of the pubKey
func (out *TxOutput) CanBeUnlocked(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTxOutput create a new TxOutput
func NewTxOutput(value int, address string) *TxOutput {
	out := &TxOutput{
		Value:      value,
		PubKeyHash: wallet.PublicKeyHash([]byte(address)),
	}

	return out
}
