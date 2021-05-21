package blockchain

import (
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
		LogErrHandle(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTX(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis Created")

		err := txn.Set(genesis.Hash, genesis.Serialize())
		if err != nil {
			LogErrHandle(err)
		}

		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})
}

func OldInitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		LogErrHandle(err)
	}

	if err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			if err = txn.Set(genesis.Hash, genesis.Serialize()); err != nil {
				LogErrHandle(err)
			}
			err = txn.Set([]byte("lh"), genesis.Hash)

			lastHash = genesis.Hash

			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				LogErrHandle(err)
			}
			err = item.Value(func(val []byte) error {
				lastHash = val

				return nil
			})
			if err != nil {
				LogErrHandle(err)
			}
			return err
		}
	}); err != nil {
		LogErrHandle(err)
	}

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	if err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			LogErrHandle(err)
		}
		if err = item.Value(func(val []byte) error {
			lastHash = val

			return nil
		}); err != nil {
			LogErrHandle(err)
		}
		return err
	}); err != nil {
		LogErrHandle(err)
	}

	newBlock := CreateBlock(data, lastHash)

	if err := chain.Database.Update(func(transaction *badger.Txn) error {
		err := transaction.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			LogErrHandle(err)
		}
		err = transaction.Set([]byte("lh"), newBlock.Hash)
		if err != nil {
			LogErrHandle(err)
		}

		chain.LastHash = newBlock.Hash

		return err
	}); err != nil {
		LogErrHandle(err)
	}
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iterator := BlockChainIterator{
		CurrentHash: chain.LastHash,
		Database:    chain.Database,
	}

	return &iterator
}

func (iterator *BlockChainIterator) Next() *Block {
	var block *Block

	if err := iterator.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash)
		if err != nil {
			LogErrHandle(err)
		}

		err = item.Value(func(val []byte) error {
			block = Deserialize(val)
			return nil
		})
		if err != nil {
			LogErrHandle(err)
		}

		return err
	}); err != nil {
		LogErrHandle(err)
	}

	iterator.CurrentHash = block.PrevHash

	return block
}
