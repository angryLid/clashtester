package main

import (
	"fmt"
	"log"
	"time"

	"github.com/angrylid/clashtester/clash"
	"github.com/angrylid/clashtester/speedtest"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// cmd.Execute()
	// panic(1)
	err := clash.SetMode(clash.GlobalMode)
	if err != nil {
		log.Fatal(`Cannot connect to clash external controller.`)
	}

	fmt.Println(`Preparing, please wait`)

	nodeList := clash.GetNodeList()

	var index = 1
	var length = len(nodeList)
	for _, name := range nodeList {
		fmt.Printf("(%d/%d) \"%s\" ", index, length, name)
		index += 1
		clash.SwitchToNode(name)
		time.Sleep(time.Second)
		// Start Test

		server, err := speedtest.SelectNearestServer()
		if err != nil {
			fmt.Printf("Timeout.\n")
			continue
		}
		totalLatency, latency, err := server.GetPingLatency()

		if err != nil {
			fmt.Printf("Timeout.\n")
			continue
		}
		fmt.Printf("TCP:%s, HTTP:%s", totalLatency, latency)
		r, err := server.MakeDownload()
		if err != nil {
			fmt.Printf("Timeout.\n")
			continue
		}
		speed, _ := r.GetSpeed()
		fmt.Printf("Down:%f", speed)
		r, err = server.MakeUpload()
		if err != nil {
			fmt.Printf("Timeout.\n")
			continue
		}
		speed, _ = r.GetSpeed()
		fmt.Printf("Up:%f", speed)

		// End Test

	}

	err = clash.SetMode(clash.RuleMode)

	if err != nil {
		fmt.Println(`You probably have to switch to rule mode manually.`)
	}

	fmt.Println("Finished.")
}
