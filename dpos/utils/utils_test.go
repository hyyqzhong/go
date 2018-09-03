package utils
import (
	"testing"
)

func TestGetEnvValue(t *testing.T) {
	//LoadingEnv("dpos.env")
	port:=""//GetNodeIdValue("PORT")
	if e := port; e == "" { //try a unit test on function
		t.Error("Not exist dpos port.") // 如果不是如预期的那么就报错
	} else {
		t.Log("dpos port get pass.", e) //记录一些你期望记录的信息
	}
}

