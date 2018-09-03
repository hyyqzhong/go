package blockchain

import (
	"bytes"
	"encoding/gob"
	"dpos/database"
	"log"
	"github.com/boltdb/bolt"
	"dpos/utils"
	"sort"
)

type Delegate struct {
	Address    string
	LastHeight int64
	NumPeer    int
	VoteNums    float64
	IsPow bool
	Votes       map[string]Coiner
}

type DelegateSlice []*Delegate

func (s DelegateSlice) Len() int { return len(s) }
func (s DelegateSlice) Swap(i, j int){ s[i], s[j] = s[j], s[i] }
func (s DelegateSlice) Less(i, j int) bool { return s[i].VoteNums < s[j].VoteNums }

//根据投票数排序候选人
func SortDelegateByVotes(delegates[] *Delegate) []*Delegate{
	delegateSlice:=DelegateSlice(delegates)
	sort.Stable(delegateSlice)
	return delegates
}


//新增或者更新候选人，是否受托人状态不更新
func AddDelegate(blockchain *BlockChain, delegate *Delegate,lastHeight int64) bool{
	isAdd:=false
	err := blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		addressData := []byte(delegate.Address)
		//查询数据存在否
		dbDelegateData := bucket.Get(addressData)
		if dbDelegateData != nil {
			dbDelegate:=DeserializeDelegate(dbDelegateData)
			if(dbDelegate.LastHeight<lastHeight){
				delegate.LastHeight=lastHeight
				delegate.IsPow=dbDelegate.IsPow//是否受托人状态不更新
				isAdd=true
			}
			return nil
		}else{
			isAdd=true
		}
		if isAdd{
			err := bucket.Put(addressData, utils.Serialize(delegate))
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return isAdd
}

//更新受托人（状态）
func UpdateDelegate(blockchain *BlockChain, delegate *Delegate) {
	err := blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		addressData := []byte(delegate.Address)
		//查询数据存在否
		dbDelegateData := bucket.Get(addressData)
		if dbDelegateData != nil {
			err := bucket.Put(addressData, utils.Serialize(delegate))
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//删除某个受托人
func DelDelegate(blockchain *BlockChain, address []byte) {
	err := blockchain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		err := bucket.Delete(address)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//获取单个受托人
func GetDelegate(blockchain *BlockChain, address []byte) *Delegate {
	var delegate Delegate
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		//查询数据存在否
		dbDelegateData := bucket.Get(address)
		if dbDelegateData != nil {
			return nil
		}
		delegate = *DeserializeDelegate(dbDelegateData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &delegate
}

//获取所有受托人
func GetAllDelegates(blockchain *BlockChain) []*Delegate {
	var delegates []*Delegate
	err := blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			delegate := DeserializeDelegate(v)
			delegates=append(delegates,delegate)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return delegates
}

//获取受托人总数
func GetNumberDelegates(blockchain *BlockChain) int{
	numberDelegate := 0
	blockchain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.DelegatesBucket))
		cursor := bucket.Cursor()
		for k,_ := cursor.First();k!=nil;k,_=cursor.Next(){
			numberDelegate += 1
		}
		return nil
	})
	return numberDelegate
}


//反序列化对象
func DeserializeDelegate(encoderDelegate []byte) *Delegate {
	var delegate Delegate
	decode := gob.NewDecoder(bytes.NewReader(encoderDelegate))
	err := decode.Decode(&delegate)
	if err != nil {
		log.Panic(err)
	}
	return &delegate
}
