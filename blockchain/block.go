package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"

	"github.com/EgorKurito/TokenCoin/util"
)

const defaultDifficult = 10

type Block struct {
	Timestamp     int64
	Hash          []byte
	Transactions  []*Transaction
	PrevBlockHash []byte
	Nonce         int
	Difficult     int
}

func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{
		time.Now().Unix(),
		[]byte{},
		txs,
		prevHash,
		0,
		defaultDifficult,
	}

	hashTx := block.HashTransactions()

	pow := NewProofOfWork()
	nonce, hash := pow.Run(prevHash, hashTx)

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer

	encoder := gob.NewEncoder(&res)
	if err := encoder.Encode(b); err != nil {
		util.LogErrHandle(err)
	}

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&block); err != nil {
		util.LogErrHandle(err)
	}

	return &block
}

// HashTransactions returns a hash of the transactions in the block
// ToDo: add merkleTree
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
