package database

import (
	"fmt"

	"github.com/boltdb/bolt"
)

const dbFile = "dpos_blockchain_%s.db"
const BlocksBucket = "dpos_blocks"
const DelegatesBucket = "dpos_delegate"
const TransfersBucket="dpos_transfer"
const LastHash="lastHash"

/*
 * @Auther: zhongyq
 * @Description: 初始化本地数据库
 */
func InitDB(nodeID string) (*bolt.DB, error) {
	dbFile := fmt.Sprintf(dbFile,nodeID)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}

	//区块桶
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BlocksBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})

	//受托人桶
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(DelegatesBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})

	//交易桶
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(TransfersBucket))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})
	return db, nil
}
