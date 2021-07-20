package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/EgorKurito/TokenCoin/blockchain/wallet"
	"github.com/EgorKurito/TokenCoin/util"
)

const reward = 200

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func NewCoinbaseTX(toAddress, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", toAddress)
	}
	txIn := NewTxInput(NewOutPoint(&[]byte{}, -1), nil, []byte(data))
	txOut := NewTxOutput(reward, toAddress)

	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{*txIn},
		Outputs: []TxOutput{*txOut},
	}

	return &tx
}

func (tx *Transaction) setID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	if err := encoder.Encode(tx); err != nil {
		util.LogErrHandle(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// IsCoinbase check transaction
// ToDo: method PreviousOutPoint.Hash.IsEqual
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].PreviousOutPoint.ID == -1
}

func NewTransaction(from *wallet.Wallet, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from.Address(), amount)

	if acc < amount {
		log.Panic("Error: Not enough funds!")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			util.LogErrHandle(err)
		}

		for _, out := range outs {
			input := TxInput{*NewOutPoint(&txID, out), from.PublicKey, nil}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTxOutput(amount, to))

	if acc < amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from.GetStringAddress()))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.setID()

	return &tx
}
