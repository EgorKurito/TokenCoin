package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/EgorKurito/TokenCoin/blockchain/wallet"
	"github.com/EgorKurito/TokenCoin/util"
)

const reward = 200

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) StringID() string {
	return string(tx.ID)
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

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Inputs {
		if _, exists := prevTXs[vin.PreviousOutPoint.StringHash()]; !exists {
			util.LogErrHandle(fmt.Errorf("ERROR: Previous transaction is not correct"))
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, vin := range tx.Inputs {
		prevTx := prevTXs[vin.PreviousOutPoint.StringHash()]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKeyHash = prevTx.Outputs[vin.PreviousOutPoint.ID].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKeyHash)
		x.SetBytes(vin.PubKeyHash[:(keyLen / 2)])
		y.SetBytes(vin.PubKeyHash[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Inputs[inId].PubKeyHash = nil
	}

	return true
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
// set sign & pubkey nil
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, vin := range tx.Inputs {
		inputs = append(inputs, *NewTxInput(&vin.PreviousOutPoint, nil, nil))
	}
	for _, vout := range tx.Outputs {
		outputs = append(outputs, TxOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

// IsCoinbase check transaction
// ToDo: method PreviousOutPoint.Hash.IsEqual
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && tx.Inputs[0].PreviousOutPoint.ID == -1
}

// NewTransaction creates a new transaction
func NewTransaction(from *wallet.Wallet, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyHash(from.PublicKey)
	acc, validOutputs := chain.FindSpendableOutputs(pubKeyHash, amount)

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

	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from.GetStringAddress()))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.setID()

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
