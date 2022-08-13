package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/angrylid/clashtester/clash"
	"github.com/angrylid/clashtester/cloudflare"
	"github.com/xuri/excelize/v2"
)

var chanRecord = make(chan cloudflare.Record)
var chanErr = make(chan error, 32)
var chanStatus = make(chan string)

func main() {
	fmt.Println("Clash Tester(rev2) is running.")
	go ErrHandler(chanErr, true)
	go RecordHandler(chanRecord, chanStatus)

	clash.SetMode(clash.GlobalMode)
	defer clash.SetMode(clash.RuleMode)
	proxy := clash.ReadConfig().Proxy

	nodeList := clash.GetFilteredNodeList(os.Args[1:])
	for _, node := range nodeList {

		clash.SwitchToNode(node)
		cf := cloudflare.New(proxy)

		r1 := make(cloudflare.LatencySlice, 0, 10)
		r2 := make(cloudflare.SpeedSlice, 0, 10)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*12)
		defer cancel()
		for range [8]bool{} {
			tracer, err := cf.Download(ctx, 100_000)
			if err != nil {
				chanErr <- err
			} else {
				r1 = append(r1, tracer.Latency())
				r2 = append(r2, tracer.Speed())
			}
		}

		r3 := make(cloudflare.SpeedSlice, 0, 10)
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*8)
		defer cancel()
		for range [8]bool{} {
			speed, err := cf.Upload(ctx, 100_000)
			if err != nil {
				chanErr <- err
			} else {
				r3 = append(r3, *speed)
			}
		}

		r4 := make(cloudflare.SpeedSlice, 0, 10)
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*40)
		defer cancel()
		for range [2]bool{} {
			tracer, err := cf.Download(ctx, 10_000_000)
			if err != nil {
				chanErr <- err
			} else {
				r4 = append(r4, tracer.Speed())
			}
		}
		chanRecord <- cloudflare.Record{
			NodeName: node,
			Latency:  r1,
			Dn100KB:  r2,
			Up100KB:  r3,
			Dn10MB:   r4,
		}
	}
	close(chanRecord)

	fileName := <-chanStatus
	fmt.Printf("finished. result is %s Press [Enter] to quit.\n", fileName)

	var input string
	fmt.Scanln(&input)

}

func ErrHandler(ch chan error, suppress bool) {
	for {
		v := <-ch
		if !suppress {
			log.Println(v)
		}
	}
}

func RecordHandler(ch chan cloudflare.Record, chStatus chan string) {
	// xlsx prepare
	sheet := "Sheet1"
	f := excelize.NewFile()
	f.SetCellValue(sheet, "A1", "Name")
	f.SetCellValue(sheet, "B1", "Latency")
	f.SetCellValue(sheet, "C1", "Dn100KB")
	f.SetCellValue(sheet, "D1", "Up100KB")
	f.SetCellValue(sheet, "E1", "Dn10MB")
	f.SetCellValue(sheet, "F1", "OK-Dn10MB")
	f.SetCellValue(sheet, "G1", "OK-Dn10MB")
	f.SetCellValue(sheet, "H1", "OK-Dn10MB")
	f.SetCellValue(sheet, "I1", "OK-Dn10MB")
	f.SetColWidth(sheet, "A", "A", 32)
	f.SetColWidth(sheet, "B", "E", 16)
	row := 2

	for {
		v, ok := <-ch
		if !ok {
			break
		}
		log.Println(v)
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), v.NodeName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), v.Latency.Avg().Cell())
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), v.Dn100KB.Avg().Cell())
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), v.Up100KB.Avg().Cell())
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), v.Dn10MB.Avg().Cell())

		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), len(v.Latency))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), len(v.Dn100KB))
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), len(v.Up100KB))
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), len(v.Dn10MB))
		row++
	}

	var fileName = func() string {
		now := strings.Replace(time.Now().Local().Format(time.RFC3339), ":", "-", -1)
		return fmt.Sprintf("Report-%s.xlsx", now[:10]+"_"+now[11:19])
	}()

	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}

	chStatus <- fileName
}
