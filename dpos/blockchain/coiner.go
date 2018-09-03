package blockchain

import (
	"dpos/utils"
	"fmt"
	"sync"
	"strconv"
	"math/rand"
)

type Coiner struct {
	Address string
	CoinNum     float64
	CandAddress string
	IsVote      bool
}

//初始化持币人
func InitCoiners(no int) []Coiner {
	fmt.Println("start init Coiners")
	var coiners []Coiner
	var coinNum float64 =1.11
	for i := 0; i < no; i++ {
		coiner := Coiner{strconv.FormatInt(int64(i),10), utils.Decimal(coinNum * 10), "", false}
		coinNum++
		coiners = append(coiners, coiner)
	}
	fmt.Println("finish init Coiners len is ", len(coiners))
	return coiners
}

//持币人进行投票
func AddVote(coiner Coiner, delegates []*Delegate, ranVoteIndex int, m *sync.Mutex) *Coiner {
	if ranVoteIndex < len(delegates) {
		m.Lock()
		curCandidate := delegates[ranVoteIndex]
		curCandidate.VoteNums += coiner.CoinNum
		curCandidate.Votes[coiner.Address] = coiner
		delegates[ranVoteIndex] = curCandidate
		coiner.CandAddress = curCandidate.Address
		coiner.IsVote = true
		m.Unlock()
	}
	return &coiner
}

func CancelVote(coiner *Coiner,delegate *Delegate,m *sync.Mutex) * Delegate{
	m.Lock()
	delegate.VoteNums-=coiner.CoinNum
	m.Unlock()
	return delegate
}


//循环投票和取消投票
func AddAndCancelVote(coiners [] Coiner,blockChain *BlockChain,m *sync.Mutex){
	delegates:=GetAllDelegates(blockChain)
	//开始进行投票
	for i, v := range coiners {
		//未投票，并且候选人数大于最低限制时启动
		if(len(delegates)>=utils.LimitDeledateNum&&!v.IsVote){
			//随机投票
			ranVoteIndex := rand.Intn(len(delegates))
			coiner := AddVote(v, delegates, ranVoteIndex, m)
			//更新投票数
			UpdateDelegate(blockChain,delegates[ranVoteIndex])
			if coiner.IsVote {
				v.IsVote = coiner.IsVote
				v.CandAddress = coiner.CandAddress
				coiners[i] = v
			}
		}
	}

	//随机挑选10个投票者取消投票
	for i:=0;i<10;i++{
		roteIndex := rand.Intn(len(coiners))
		curCoiner:=coiners[roteIndex]
		if curCoiner.IsVote{
			delegate:=GetDelegate(blockChain,[]byte(curCoiner.CandAddress))
			if delegate!=nil{
				delegate=CancelVote(&curCoiner,delegate,m)
				//更新投票数
				UpdateDelegate(blockChain,delegate)
				curCoiner.IsVote=false
				curCoiner.CandAddress=""
				coiners[roteIndex]=curCoiner
			}
		}
	}
}
