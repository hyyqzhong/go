package blockchain

import (
	"github.com/boltdb/bolt"
	"log"
	"dpos/database"
	"bytes"
	"encoding/gob"
	"dpos/utils"
)

type Transfer struct {
	TranId     string
	From       string
	To         string
	TranAmount float64
	TransferBy string
}

//新增交易
func AddTransfer(db *bolt.DB,transfer *Transfer){
	err:=db.Update(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(database.TransfersBucket))
		tranId:=[]byte(transfer.TranId)
		//查询数据存在否
		dbTransferData := bucket.Get(tranId)
		if dbTransferData != nil {
			return nil
		}
		err := bucket.Put(tranId, utils.Serialize(transfer))
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
}

//删除某个交易
func DelTransfer(db *bolt.DB,tranId []byte){
	err:=db.Update(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(database.TransfersBucket))
		err:=bucket.Delete(tranId)
		if err != nil {
			log.Panic(err)
		}
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
}

//获取单个交易
func GetTransfer(db *bolt.DB,tranId []byte) *Transfer {
	var transfer Transfer
	err:=db.View(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(database.TransfersBucket))
		//查询数据存在否
		dbTransferData := bucket.Get(tranId)
		if dbTransferData != nil {
			return nil
		}
		transfer=*DeserializeTransfer(dbTransferData)
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	return &transfer
}

//获取所有交易
func GetAllTransfers(db *bolt.DB) []Transfer {
	var transfers []Transfer
	err:=db.View(func(tx *bolt.Tx) error {
		bucket:=tx.Bucket([]byte(database.TransfersBucket))
		cursor:=bucket.Cursor()
		var i=0
		for k,v:=cursor.First();k!=nil;k,v=cursor.Next(){
			transfer:=*DeserializeTransfer(v)
			transfers[i]=transfer
			i++
		}
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	return transfers
}

//反序列化对象
func DeserializeTransfer(encoderTransfer []byte) *Transfer {
	var transfer Transfer
	decode:=gob.NewDecoder(bytes.NewReader(encoderTransfer))
	err:=decode.Decode(&transfer)
	if err!=nil{
		log.Panic(err)
	}
	return &transfer
}


//计算markel树hash
func CalTransHash(transfers []Transfer) string{
	if(len(transfers)>0){
		var hashs []string
		for i,transfer:= range transfers{
			hashs[i]=calTransferHash(transfer)
		}
		tempHashs:=iteratorTranHash(hashs)
		return tempHashs[0]
	}
	return utils.CalculateHash("")
}


func calTransferHash(transfer Transfer) string{
	record:=transfer.TranId+transfer.From+transfer.To+utils.FloatToStr(transfer.TranAmount)+transfer.TransferBy
	return utils.CalculateHash(record)
}

//递归计算，直至hashs的长度为1
func iteratorTranHash(hashs []string) []string{
	if(len(hashs)==1){
		return hashs
	}
	var tempHashs []string
	j:=0
	for i:=0;i<len(hashs);i++{
		if i%2==1 {//两两合并
			tempHashs[j]=utils.CalculateHash(hashs[i]+hashs[i-1])
			j++
		}else if i==(len(hashs)-1){//说明单数
			tempHashs[j]=hashs[i]
			break
		}
	}
	if(len(tempHashs)>1) {
		tempHashs = iteratorTranHash(tempHashs)
	}
	return tempHashs
}
