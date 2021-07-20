package blockchain

import (
	"github.com/EgorKurito/TokenCoin/util"
	"github.com/dgraph-io/badger"
)

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
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
