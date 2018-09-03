package utils

import (
	"fmt"
	"strconv"
	"github.com/joho/godotenv"
	"os"
	"log"
	"crypto/sha256"
	"encoding/hex"
	"bytes"
	"encoding/gob"
)
type UUID [16]byte



//保留两位小数浮点数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}


func LoadingEnv(fileName string) {
	err:=godotenv.Load("dpos.env")
	if err!=nil{
		log.Fatal(err)
	}
}

func GetEnvValue(key string) string{
	return os.Getenv(key)
}

// SHA256 计算方法
func CalculateHash(s string) string{
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//将string转成固定长度数组，前提改str要小于该数组，不然会报错
func ConvertStrToBytes(str string) []byte{
	var bytes [constLength]byte
	for i,c := range str{
		bytes[i] = byte(c)
	}
	return bytes[:]
}

//解析定长数组，空值去掉
func ConvertBytesToStr(tempBytes []byte) string{
	var temp []byte
	for _,b:=range tempBytes{
		if b!=0x0{
			temp=append(temp,b)
		}
	}
	return fmt.Sprintf("s%",temp)
}


//序列化对象
func  Serialize(data interface{}) []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}


func FloatToStr(v float64)string{
	return strconv.FormatFloat(v, 'E', -1, 64)
}

func SubString(str string,begin,length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)
	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}
	// 返回子串
	return string(rs[begin:end])
}