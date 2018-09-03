package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"dpos/utils"
)

type Block struct {
	//区块头部
	Height    int64 //最大高度
	CurHash   string //当前hash
	PreHash   string //前一个区块hash
	Timestamp int64  //时间戳
	TransMarkelHash string //交易markel树
	Difficulty int64 //困难度
	Nounce    int64 //难易程度

	//区块内容
	Transfers []Transfer //交易内容
}


// 反序列化区块
func DeserializeBlock(encodedBlock []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(encodedBlock))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}


//计算区块当前hash
func CaculateBlockHash(block *Block) string{
	//前面固定，后面3个是变化的
	record:=string(block.Height)+block.PreHash+string(block.Difficulty)+block.TransMarkelHash+string(block.Timestamp)+string(block.Nounce)
	return utils.CalculateHash(record)
}