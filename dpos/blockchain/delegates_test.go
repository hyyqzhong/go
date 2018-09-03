package blockchain

import (
	"testing"
	"dpos/database"
	"log"
)

func TestAddDelegate(t *testing.T) {

	NodeAddress="local:3000"
	//连接数据库，创建创世区块（如果已经有不创建）
	db, err := database.InitDB("3000")
	if err != nil {
		log.Panic(err)
	}
	blockChain := &BlockChain{db, []byte("")}

	delegate := &Delegate{NodeAddress, 0, 0,0,false,make(map[string] Coiner)}
	AddDelegate(blockChain, delegate,0)
	dbDelegate:=GetDelegate(blockChain,[]byte(NodeAddress))

	if e := dbDelegate; e == nil{ //try a unit test on function
		t.Error("Not exist Delegate.") // 如果不是如预期的那么就报错
	} else {
		t.Log("AddDelegate get pass.", e) //记录一些你期望记录的信息
	}
}

func TestSortDelegateByVotes(t *testing.T) {

	NodeAddress="local:3000"

	delegate1 := &Delegate{NodeAddress, 0, 0,2,false,make(map[string] Coiner)}
	delegate2 := &Delegate{NodeAddress, 0, 0,1,false,make(map[string] Coiner)}
	delegates:=[]*Delegate{delegate1,delegate2}
	SortDelegateByVotes(delegates)

	if e := delegates; e[0].VoteNums != 1{ //try a unit test on function
		t.Error("SortDelegateByVotes fail.") // 如果不是如预期的那么就报错
	} else {
		t.Log("SortDelegateByVotes  pass.", e) //记录一些你期望记录的信息
	}
}
