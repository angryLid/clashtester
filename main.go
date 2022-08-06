package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/angrylid/clashtester/clash"
	"github.com/angrylid/clashtester/cloudflare"
	"github.com/xuri/excelize/v2"
)

func main() {
	clash.SetMode(clash.GlobalMode)
	defer clash.SetMode(clash.RuleMode)
	nodeList := clash.GetFilteredNodeList(os.Args[1:])

	proxy := clash.ReadConfig().Proxy
	writeLine, save := Excel()
	for _, node := range nodeList {
		client := cloudflare.New(proxy)

		clash.SwitchToNode(node)

		var record cloudflare.Record

		d100k := make(cloudflare.TracerList, 0, 50)
		for range [10]struct{}{} {
			tracer, err := client.Download(100_000)
			if err != nil {
				continue
			}
			d100k = append(d100k, tracer)
		}

		record.Latency = d100k.AvgLatency()
		record.Dn100KB = d100k.AvgSpeed()

		tracer, err := client.Download(25_000_000)
		if err != nil {
			record.Dn25MB = cloudflare.NaS()
		} else {
			record.Dn25MB = tracer.Speed()

		}
		u100k := cloudflare.Speed{}
		for range [10]struct{}{} {
			speed, err := client.Upload(100_1000)
			if err != nil {
				continue
			}
			u100k = u100k.Add(*speed)
		}
		record.Up100KB = u100k
		log.Println(node, record)
		writeLine(node, record)
	}
	save()
}

func Excel() (func(name string, record cloudflare.Record) error, func()) {
	sheet := "Sheet1"
	f := excelize.NewFile()
	f.SetCellValue(sheet, "A1", "Name")
	f.SetCellValue(sheet, "B1", "Latency")
	f.SetCellValue(sheet, "C1", "Dn100KB")
	f.SetCellValue(sheet, "D1", "Dn25MB")
	f.SetCellValue(sheet, "E1", "Up100KB")
	f.SetColWidth(sheet, "A", "A", 32)
	f.SetColWidth(sheet, "B", "E", 16)

	row := 2

	var writeLine = func(name string, record cloudflare.Record) error {
		var err error
		err = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), name)
		if err != nil {
			return err
		}
		err = f.SetCellValue(sheet, fmt.Sprintf("B%d", row),
			record.Latency.String()[:len(record.Latency.String())-2])
		if err != nil {
			return err
		}
		err = f.SetCellValue(sheet, fmt.Sprintf("C%d", row),
			fmt.Sprintf("%.2f", record.Dn100KB.Value()))
		if err != nil {
			return err
		}
		err = f.SetCellValue(sheet, fmt.Sprintf("D%d", row),
			fmt.Sprintf("%.2f", record.Dn25MB.Value()))
		if err != nil {
			return err
		}
		err = f.SetCellValue(sheet, fmt.Sprintf("E%d", row),
			fmt.Sprintf("%.2f", record.Up100KB.Value()))
		if err != nil {
			return err
		}
		row++
		return nil
	}
	var save = func() {
		var fileName = func() string {
			now := strings.Replace(time.Now().Local().Format(time.RFC3339), ":", "-", -1)
			return now[:10] + "_" + now[11:19]
		}

		if err := f.SaveAs(fmt.Sprintf("Report-%s.xlsx", fileName())); err != nil {
			fmt.Println(err)
		}
	}

	return writeLine, save
}
