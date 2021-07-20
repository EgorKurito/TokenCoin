package blockchain

import (
	"bytes"

	"github.com/EgorKurito/TokenCoin/blockchain/wallet"
)

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
	PubKeyHash       []byte
	Signature        []byte
}

// NewTxInput create a new TxInput
func NewTxInput(prevOut *OutPoint, sign, pubKey []byte) *TxInput {
	newTxInput := &TxInput{
		PreviousOutPoint: *prevOut,
		PubKeyHash:       pubKey,
		Signature:        sign,
	}

	return newTxInput
}

func (in *TxInput) CanUnlock(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKeyHash)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
