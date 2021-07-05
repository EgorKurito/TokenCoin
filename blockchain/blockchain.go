package blockchain

import (
	"egorkurito/TokenCoin/util"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger" // This is our database import
)

const (
	dbPath = "./tmp/blocks"
	dbFile = "./tmp/blocks/MANIFEST"

	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func DBExists(db string) bool {
	if _, err := os.Stat(db); os.IsNotExist(err) {
		return false
	}
	return true
}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBExists(dbFile) {
		fmt.Println("blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		util.LogErrHandle(err)
	}

	if err := db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTX(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis Created")

		if err := txn.Set(genesis.Hash, genesis.Serialize()); err != nil {
			util.LogErrHandle(err)
		}

		if err := txn.Set([]byte("lh"), genesis.Hash); err != nil {
			util.LogErrHandle(err)
		}

		lastHash = genesis.Hash

		return err
	}); err != nil {
		util.LogErrHandle(err)
	}

	chain := BlockChain{lastHash, db}
	return &chain
}

func ContinueBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBExists(dbFile) == false {
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

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
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
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXNs[inTxID] = append(spentTXNs[inTxID], in.Out)
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

func (chain *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(address)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
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

func (iterator *BlockChainIterator) Next() *Block {
	var block *Block

	if err := iterator.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash)
		if err != nil {
			util.LogErrHandle(err)
		}

		err = item.Value(func(val []byte) error {
			block = Deserialize(val)
			return nil
		})
		if err != nil {
			util.LogErrHandle(err)
		}

		return err
	}); err != nil {
		util.LogErrHandle(err)
	}

	iterator.CurrentHash = block.PrevBlockHash

	return block
}
