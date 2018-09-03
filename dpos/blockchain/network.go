package blockchain

import (
	"fmt"
	"log"
	"net"
	"time"
	"dpos/utils"
	"io/ioutil"
	"io"
	"bytes"
	"encoding/gob"
)

const protocol = "tcp"
const constLength = 12

var NodeAddress string
var DelegateFlag= false
var blockFlag=true

func StartServer(nodeAddress string, blockChain *BlockChain) {

	log.Println("start server node address is ", nodeAddress)
	listen, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer listen.Close()

	//判断当前节点，是否受托人节点，如果是则开启出块权利
	go Fork(blockChain)


	//接收广播消息，出块、新增或更新受托人、新增交易
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, blockChain)
	}

}

//接收连接，需要处理几类：广播交易数据、新增的受托人、验证区块
func handleConnection(conn net.Conn, blockChain *BlockChain) {
	//创建连接后再开始接收连接
	time.Sleep(time.Second)
	var bussinessType = ""
	switch bussinessType {
	case utils.BlockChan:
		handleBlock(conn, blockChain)
	case utils.DelegateChan:
		handleDelegate(conn, blockChain)
	case utils.TransferChan:
		handleTransfer(conn, blockChain)
	case utils.DelegateFlag:
		handleDelegateFlag(conn, blockChain)
	default:
		log.Println("unknown connect type")
	}
}

//接收广播过来的区块，并处理(不包括每条交易自验证)
func handleBlock(conn net.Conn, blockChain *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	blockData := utils.Serialize(request[constLength:])
	block := DeserializeBlock(blockData)
	//这里面要涉及到停止自己出块计算工作
	if DelegateFlag {
		blockFlag=false//收到广播的区块后，停止自己出块功能
		//验证区块,忽略每条交易记录正确性验证（数字签名等验证）
		isNormal:=VerfyBlock(blockChain.GetLastBlock(),block)
		if isNormal{//区块验证通过，添加到本地区块，并广播该区块
			//将收到的区块添加进本地
			blockChain.AddBlock(block)
			log.Println("Add block by receive, hash is ", block.CurHash)
			//删除相关交易
			for _,transfer:=range block.Transfers{
				DelTransfer(blockChain.db,[]byte(transfer.TranId))
			}
			//广播区块至别的受托人
			SendDelegatesBlock(blockChain,block)
		}else{
			log.Println("Verfy block fail, hash is ", block.CurHash)
		}
		blockFlag=true
	}
}

//接收受托人状态
func handleDelegateFlag(conn net.Conn, blockChain *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	blockData := utils.Serialize(request[constLength:])
	DelegateFlag= deserializeFlag(blockData)
}

//接收新添加的交易
func handleTransfer(conn net.Conn, blockChain *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	tempData := utils.Serialize(request[constLength:])
	transfer := DeserializeTransfer(tempData)
	//本地添加交易
	AddTransfer(blockChain.db,transfer)
	//广播交易
	SendDelegatesTransfer(blockChain,transfer)
}

//接收新添加的受托人
func handleDelegate(conn net.Conn, blockChain *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	tempData := utils.Serialize(request[constLength:])
	delegate := DeserializeDelegate(tempData)
	lastHeight := blockChain.GetLastHeight()
	//保存新受托人
	isAdd:=AddDelegate(blockChain,delegate,lastHeight)

	//广播出去
	if isAdd{
		SendDelegates(blockChain,delegate)
	}
}


//发送受托人至其他节点
func SendDelegates(blockchain *BlockChain,delegate *Delegate){
	delegates:=GetAllDelegates(blockchain)
	for _,dbDelegate:=range delegates{
		if dbDelegate.Address!=delegate.Address{
			SendDelegate(dbDelegate.Address,delegate)
		}
	}
}

//发送区块至其他节点
func SendDelegatesBlock(blockchain *BlockChain,block *Block){
	delegates:=GetAllDelegates(blockchain)
	for _,dbDelegate:=range delegates{
		if dbDelegate.Address!=NodeAddress&&dbDelegate.IsPow{
			SendBlock(dbDelegate.Address,block)
		}
	}
}

//发送交易至其他节点
func SendDelegatesTransfer(blockchain *BlockChain,transfer *Transfer){
	delegates:=GetAllDelegates(blockchain)
	for _,dbDelegate:=range delegates{
		if dbDelegate.Address!=NodeAddress{
			SendTransfer(dbDelegate.Address,transfer)
		}
	}
}


//发送受托人至指定节点
func SendDelegate(addressUrl string, delegate *Delegate) {
	tempData := utils.Serialize(delegate)
	request := append(utils.ConvertStrToBytes(utils.DelegateChan), tempData...)
	sendData(addressUrl, request)
}

//发送交易至指定节点
func SendTransfer(addressUrl string, transfer *Transfer) {
	tempData := utils.Serialize(transfer)
	request := append(utils.ConvertStrToBytes(utils.BlockChan), tempData...)
	sendData(addressUrl, request)
}

//发送区块至指定节点
func SendBlock(addressUrl string, block *Block) {
	tempData := utils.Serialize(block)
	request := append(utils.ConvertStrToBytes(utils.BlockChan), tempData...)
	sendData(addressUrl, request)
}


//发送候选人状态至指定节点
func SendDelegateFlag(addressUrl string, tempFlag bool) {
	tempData := utils.Serialize(tempFlag)
	request := append(utils.ConvertStrToBytes(utils.DelegateFlag), tempData...)
	sendData(addressUrl, request)
}

//每隔一秒生成交易
func GenTransfer(blockchain *BlockChain){
	for{
		time.Sleep(time.Second)
		transfer:=Transfer{utils.GetUuid(),"zhangsan","lisi",float64(0.00),NodeAddress}
		fmt.Println("genTransfer is ",transfer)
		SendDelegatesTransfer(blockchain,&transfer)
	}
}


//循环出块
func Fork(blockChain *BlockChain){
	for{
		time.Sleep(time.Second*3)//每隔3秒进行一次判断
		if DelegateFlag&&blockFlag {
			//计算出块
			transfers:=GetAllTransfers(blockChain.db)
			block:=newBlock(blockChain.GetLastBlock(),transfers,blockFlag)
			if  block!=nil{
				//将区块增加到本地
				blockChain.AddBlock(block)

				//点播区块
				SendDelegatesBlock(blockChain,block)
			}
		}
	}
}



//发送数据至指定节点
func sendData(addressUrl string, data []byte) {
	conn, err := net.Dial(protocol, addressUrl)
	_, err = io.Copy(conn, bytes.NewBuffer(data))
	defer conn.Close()
	if err != nil {
		log.Panic(err)
	}
}


// 反序列化区块
func deserializeFlag(encodedBlock []byte)  bool {
	var temp bool
	decoder := gob.NewDecoder(bytes.NewReader(encodedBlock))
	err := decoder.Decode(&temp)
	if err != nil {
		log.Panic(err)
	}
	return temp
}
