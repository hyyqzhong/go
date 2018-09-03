package main

import (
	"fmt"

	"sync"
	"dpos/utils"
	"dpos/blockchain"
	"dpos/database"
	"log"
	"time"
)

var coiners []blockchain.Coiner


func main() {
	utils.LoadingEnv("dpos.env")
	port := utils.GetEnvValue("PORT")
	blockchain.NodeAddress = fmt.Sprintf("localhost:%s", port)
	log.Println("start server node address is ", blockchain.NodeAddress)

	//1.初始化持币者
	coiners = blockchain.InitCoiners(utils.CoinNum)

	//连接数据库，创建创世区块（如果已经有不创建）
	db, err := database.InitDB(port)
	if err != nil {
		log.Panic(err)
	}

	//2.获取最大高度区块，如果无则创建创世区块
	//TODO 这里面未考虑一开始增加节点同步区块至本地节点功能
	blockChain := blockchain.CreateGenesisBlock(db)
	lastHeight := blockChain.GetLastHeight()
	numberDelegate := blockchain.GetNumberDelegates(blockChain)
	//3.将自己添加到候选人中
	delegate := &blockchain.Delegate{blockchain.NodeAddress, lastHeight, numberDelegate, 0, false, make(map[string]blockchain.Coiner)}
	blockchain.AddDelegate(blockChain, delegate, lastHeight)

	//广播候选人其他节点
	blockchain.SendDelegates(blockChain, delegate)

	//4.每隔1秒产生一笔交易发送
	go blockchain.GenTransfer(blockChain)

	var m sync.Mutex
	//5.循环投票和取消投票
	blockchain.AddAndCancelVote(coiners, blockChain, &m)

	//6.重排序候选者，挑选出有权利出块受托人
	delegates := blockchain.GetAllDelegates(blockChain)
	blockchain.SortDelegateByVotes(delegates)
	//7.获取投票结果，并产生21个受托人
	if len(delegates) >= utils.LimitDeledateNum {
		var delegateMap map[string]*blockchain.Delegate
		//将失去受托人节点设置为失去受托人
		for i, tempDelegate := range delegates {
			if i < len(delegates)-utils.LimitDeledateNum {
				if(blockchain.NodeAddress!=tempDelegate.Address){
					go blockchain.SendDelegateFlag(tempDelegate.Address,false)
				}else{
					blockchain.DelegateFlag=false
				}
			}else{
				delegateMap[tempDelegate.Address]=tempDelegate
			}
		}

		//随机轮排受托人节点
		for _,tempDelegate:=range delegateMap{
			time.Sleep(time.Second*3)
			if(blockchain.NodeAddress!=tempDelegate.Address){
				go blockchain.SendDelegateFlag(tempDelegate.Address,true)
			}else{
				blockchain.DelegateFlag=true
			}
		}

	}else{
		log.Println("受托人，人数不够")
	}

	/*前面步骤为准备交易；准备受托人；接下来开始出块工作，
	*启动节点监听
	 */
	blockchain.StartServer(port, blockChain)

}
