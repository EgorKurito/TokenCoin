package blockchain

import (
	"bytes"
	"errors"
	"github.com/EgorKurito/TokenCoin/storage"
	"log"

	//  This is our database import
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/EgorKurito/TokenCoin/util"
	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tmp/blocks"
	dbFile = "./tmp/blocks/MANIFEST"

	genesisData = "First Transaction from NewGenesisBlock"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// InitBlockChain creates a new blockchain with genesisBlock
func InitBlockChain(address string) *BlockChain {
	var lastBlockHash []byte

	if dbExists(dbFile) {
		fmt.Println("blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		util.LogErrHandle(err)
	}

	if err := db.Update(func(txn *badger.Txn) error {
		genesis := NewGenesisBlock(address)
		fmt.Println("NewGenesisBlock Created")

		if err := txn.Set(genesis.Hash, genesis.Serialize()); err != nil {
			util.LogErrHandle(err)
		}

		if err := txn.Set([]byte("lh"), genesis.Hash); err != nil {
			util.LogErrHandle(err)
		}

		lastBlockHash = genesis.Hash

		return err
	}); err != nil {
		util.LogErrHandle(err)
	}

	chain := BlockChain{lastBlockHash, db}
	fmt.Println("CreateBlockchain Success!")

	return &chain
}

// LoadBlockChain load Blockchain
func LoadBlockChain() *BlockChain {
	var lastHash []byte

	if dbExists(dbFile) == false {
		fmt.Println("No blockchain found, please create one first")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		util.LogErrHandle(err)
	}

	if err := db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			util.LogErrHandle(err)
		}

		if err := item.Value(func(val []byte) error {
			lastHash = val

			return err
		}); err != nil {
			util.LogErrHandle(err)
		}

		return err
	}); err != nil {
		util.LogErrHandle(err)
	}

	chain := BlockChain{lastHash, db}
	return &chain
}

// AddBlock add the block into the blockchain
func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	if err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			util.LogErrHandle(err)
		}
		if err = item.Value(func(val []byte) error {
			lastHash = val

			return nil
		}); err != nil {
			util.LogErrHandle(err)
		}
		return err
	}); err != nil {
		util.LogErrHandle(err)
	}

	newBlock := CreateBlock(transactions, lastHash)

	if err := chain.Database.Update(func(transaction *badger.Txn) error {
		err := transaction.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			util.LogErrHandle(err)
		}
		err = transaction.Set([]byte("lh"), newBlock.Hash)
		if err != nil {
			util.LogErrHandle(err)
		}

		chain.LastHash = newBlock.Hash

		return err
	}); err != nil {
		util.LogErrHandle(err)
	}
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iterator := BlockChainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}

	return &iterator
}

func (chain *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTXNs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXNs[txID] != nil {
					for _, spentOut := range spentTXNs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					if in.CanUnlock(pubKeyHash) {
						inTxID := hex.EncodeToString(in.PreviousOutPoint.Hash)
						spentTXNs[inTxID] = append(spentTXNs[inTxID], in.PreviousOutPoint.ID)
					}
				}
			}
			if len(block.PrevBlockHash) == 0 {
				break
			}
		}
		return unspentTxs

	}
}

func (chain *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(pubKeyHash)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOuts
}

func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			util.LogErrHandle(fmt.Errorf("ERROR: Invalid transaction"))
		}
	}

	lastHash, err := storage.GetLastBlock(chain.Database)
	if err != nil {
		util.LogErrHandle(err)
	}

	newBlock := CreateBlock(transactions, lastHash)

	err = storage.SaveBlock(chain.Database, newBlock.Hash)
	if err != nil {
		util.LogErrHandle(err)
	}
	chain.LastHash = newBlock.Hash

	return newBlock
}

// VerifyTransaction verifies transaction input signatures
func (chain *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]Transaction)
	for _, vin := range tx.Inputs {
		prevTX, err := chain.FindTransaction(&vin.PreviousOutPoint.Hash)
		if err != nil {
			util.LogErrHandle(err)
		}
		prevTXs[prevTX.StringID()] = prevTX
	}

	return tx.Verify(prevTXs)
}

// FindTransaction finds a transaction by its ID
func (chain *BlockChain) FindTransaction(ID *[]byte) (Transaction, error) {
	chainIterator := chain.Iterator()

	for {
		block := chainIterator.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, *ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("transaction is not found")
}

func dbExists(db string) bool {
	if _, err := os.Stat(db); os.IsNotExist(err) {
		return false
	}
	return true
}
