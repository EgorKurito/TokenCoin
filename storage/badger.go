package storage

import (
	"github.com/EgorKurito/TokenCoin/util"
	"github.com/dgraph-io/badger"
)

func GetLastBlock(db *badger.DB) (lastHash []byte, err error) {
	err = db.View(func(txn *badger.Txn) error {
		lastHashItem, err := txn.Get([]byte("lh"))
		if err != nil {
			util.LogErrHandle(err)
		}
		if err = lastHashItem.Value(func(val []byte) error {
			lastHash = val

			return nil
		}); err != nil {
			util.LogErrHandle(err)
		}

		return nil
	})

	return
}

func SaveBlock(db *badger.DB,blockHash []byte)   {

}