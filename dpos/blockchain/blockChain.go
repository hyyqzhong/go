package blockchain

import (
	"github.com/boltdb/bolt"
	"dpos/database"
	"log"
	"errors"
	"dpos/utils"
	"time"
	"strings"
)

type BlockChain struct {
	db  *bolt.DB
	bls []byte
}

type BlockChainIterator struct {
	db          *bolt.DB
	currentHash []byte
}

//将区块添加进数据库桶中
func (blockChain *BlockChain) AddBlock(block *Block) {
	err := blockChain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.BlocksBucket))
		curHash := []byte(block.CurHash)
		//查询数据存在否
		dbBlockData := bucket.Get(curHash)
		if dbBlockData != nil {
			return nil
		}
		blockData := utils.Serialize(block)
		err := bucket.Put([]byte(block.CurHash), blockData)
		if err != nil {
			log.Panic(err)
		}
		//设置最大高度
		lastHash := bucket.Get([]byte(database.LastHash))
		var isPutLashHash = false
		if lastHash != nil {
			lastBlockData := bucket.Get(lastHash)
			lastBlock := *DeserializeBlock(lastBlockData)
			if lastBlock.Height < block.Height {
				isPutLashHash = true
			}
		} else {
			isPutLashHash = true
		}
		if isPutLashHash {
			err = bucket.Put([]byte(database.LastHash), curHash)
			if err != nil {
				log.Panic(err)
			}
			blockChain.bls = curHash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//根据某个区块hash获取区块信息
func (blockChain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block
	err := blockChain.db.View(func(tx *bolt.Tx) error {
		//获取区块桶
		bucket := tx.Bucket([] byte(database.BlocksBucket))
		blockData := bucket.Get(blockHash)
		if blockData == nil {
			return errors.New("this block is not exist")
		}
		block = *DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return block, err
}

//获取所有区块
func GetAllBlocks(db *bolt.DB) *[]Block {
	var blocks []Block
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.TransfersBucket))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			block := *DeserializeBlock(v)
			blocks = append(blocks,block)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &blocks
}

//获取最大区块
func (blockChain *BlockChain) GetLastBlock() Block {
	var lastBlock Block
	err := blockChain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.BlocksBucket))
		lastHash := bucket.Get([]byte(database.LastHash))
		if lastHash != nil {
			lastBlockData := bucket.Get(lastHash)
			lastBlock = *DeserializeBlock(lastBlockData)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastBlock
}

//获取最大区块高度
func (blockChain *BlockChain) GetLastHeight() int64 {
	var lastBlock = blockChain.GetLastBlock()
	return lastBlock.Height
}

//生成创世区块
func CreateGenesisBlock(db *bolt.DB) *BlockChain {
	var curHash []byte
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(database.BlocksBucket))
		lastHash := bucket.Get([]byte(database.LastHash))
		if lastHash == nil {
			block := &Block{0, "", "", time.Now().Unix(), "", 1, 0, nil}
			block.CurHash = CaculateBlockHash(block)
			curHash = []byte(block.CurHash)
			err := bucket.Put(curHash, utils.Serialize(block))
			if err != nil {
				log.Panic(err)
			}
			err = bucket.Put([]byte(database.LastHash), curHash)
			if err != nil {
				log.Panic(err)
			}
		} else {
			//返回的不是创世区块的hash，而是当前最大高度的hash
			curHash = lastHash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	tempBlockChain := BlockChain{db, curHash}
	return &tempBlockChain
}

//工作量计算hash函数出块
func newBlock(preBlock Block, transfers []Transfer,blockFlag bool) *Block {
	//忽略每条交易记录正确性验证（数字签名等验证）
	log.Println("start to caculate new Block")
	transMarkelHash := CalTransHash(transfers)
	log.Println("newBlock transMarkelHash is ",transMarkelHash)
	timestemp := time.Now().Unix()
	var nounce int64 = 0
	var block *Block

	//这儿尝试了3秒退出写法，没写成功：
	//timeout := time.After(time.Second * 10)
	//select {
	//case <-time.After(time.Second * time.Duration(3)):
	//	code
	//}

	for {
		if blockFlag{
			block = &Block{preBlock.Height + 1, "", preBlock.CurHash, timestemp, transMarkelHash, preBlock.Difficulty + 1, nounce, transfers}
			tempHash := CaculateBlockHash(block)
			if  strings.HasPrefix(tempHash, "000") { //区块工作量计算
				log.Println("current newBlock hash is ",tempHash)
				block.CurHash = tempHash
				break
			} else {
				log.Println("current  hash not correct is ",tempHash)
				timestemp = time.Now().Unix()
				nounce++
			}
		}else{
			block=nil
			break
		}
	}
	log.Println("end to caculate newBlock ",block)
	return block
}

//验证区块
func VerfyBlock(lastBlock Block,curBlock *Block) bool{
	isNormalBlock:=false
	if(lastBlock.CurHash!=curBlock.CurHash&&lastBlock.CurHash==curBlock.PreHash){//不是当前区块，并且最后一个区块是新产生区块的上一个区块
		//这里面未考虑区块同步至最新问题
		// 验证hash //忽略每条交易记录正确性验证（数字签名等验证）
		transMarkelHash:=CalTransHash(curBlock.Transfers)

		if(transMarkelHash==curBlock.TransMarkelHash){//交易markel树是正确的
			tempHash := CaculateBlockHash(curBlock)
			if(tempHash==curBlock.CurHash){//当前hash也是正确的
				isNormalBlock=true
			}
		}
	}
	return isNormalBlock
}
