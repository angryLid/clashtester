package clash

import (
	"reflect"
	"testing"
	"time"
)

func TestSetMode(t *testing.T) {
	SetMode(RuleMode)
}

func TestGetNodeList(t *testing.T) {
	nodes := GetNodeList()

	for key, value := range nodes {
		t.Log(key, value)
	}
}

func TestSwitchToNode(t *testing.T) {
	nodes := GetNodeList()

	keys := reflect.ValueOf(nodes).MapKeys()

	randomName := keys[int(time.Now().Unix())%len(keys)].Interface().(string)

	t.Log(randomName)
	SwitchToNode(randomName)
}
