package clash

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sort"
)

type Record = map[string]any

func ValidProtocol(protocol string) bool {
	return protocol == "Shadowsocks" ||
		protocol == "ShadowsocksR" ||
		protocol == "Vmess" ||
		protocol == "Trojan"
}

type Node struct {
	Name    string           `json:"name"`
	Type    string           `json:"type"`
	UDP     bool             `json:"udp"`
	History []map[string]any `json:"history"`
}
type ResponseBody struct {
	Proxies map[string]Node `json:"proxies"`
}

var axios = NewAxios()

func GetNodeList() []string {
	nodeNames := make([]string, 0, 99)
	resp, err := axios.Get("/proxies", nil)

	if err != nil {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	body := make(map[string]any)

	err = json.Unmarshal(bytes, &body)

	if err != nil {
		log.Fatal(err)
	}
	proxies := body["proxies"].(Record)

	for name, node := range proxies {
		node := node.(Record)
		if ValidProtocol(node["type"].(string)) {
			nodeNames = append(nodeNames, name)
		}
	}
	sort.Strings(nodeNames)

	return nodeNames
}

type Mode string

const (
	GlobalMode Mode = "Global"
	RuleMode   Mode = "Rule"
)

func SetMode(mode Mode) error {
	_, err := axios.Patch("/configs", map[string]interface{}{
		"mode": mode,
	})

	if err != nil {
		return err
	}

	return nil
}

func SwitchToNode(nodeName string) error {
	resp, err := axios.Put("/proxies/GLOBAL", map[string]any{
		"name": nodeName,
	})

	body := make(map[string]any)
	bytes, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(bytes, &body)

	if err != nil {
		return err
	}

	return nil
}
