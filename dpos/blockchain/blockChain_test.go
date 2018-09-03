package blockchain

import (
	"testing"
	"dpos/utils"
	"sync"
	"dpos/database"
	"log"
)

func TestAddAndCancelVote(t *testing.T){
	//1.初始化持币者
	var coiners = InitCoiners(utils.CoinNum)
	var m sync.Mutex

	//连接数据库，创建创世区块（如果已经有不创建）
	db, err := database.InitDB("30001")
	if err != nil {
		log.Panic(err)
	}

	//2.获取最大高度区块，如果无则创建创世区块
	//TODO 这里面未考虑一开始增加节点同步区块至本地节点功能
	blockChain := CreateGenesisBlock(db)

	//5.循环投票和取消投票
	AddAndCancelVote(coiners, blockChain, &m)



}
