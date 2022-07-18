package main

import (
	"errors"
	"fmt"
	"log"
	"m6kparse/capture"
	"os"
)

func help() {
	fmt.Println("Usage:", os.Args[0], " <Icon IP> <Mainframe IP> <mode> <source>")
	fmt.Println("")
	fmt.Println("mode can be:")
	fmt.Println(" -live: live capture from a network interface")
	fmt.Println(" -pcap: read from a pcap")
	fmt.Println("")
	fmt.Println("source can be:")
	fmt.Println(" In live mode, a network interface")
	fmt.Println(" In pcap mode, a pcap file")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("", os.Args[0], "192.168.1.125 192.168.1.125 -live eth0")
	fmt.Println("", os.Args[0], "192.168.1.125 192.168.1.125 -pcap /tmp/capture.pcap")
}

func main() {
	var err error
	if len(os.Args) != 5 {
		help()
		return
	}
	iconIP := os.Args[1]
	frameIP := os.Args[2]
	mode := os.Args[3]
	source := os.Args[4]

	f, _ := os.Create("output.log")
	logs := log.New(f, "M6kParser", log.Lshortfile)
	defer f.Close()

	cap := capture.New(logs, iconIP, frameIP)

	if mode == "-live" {
		err = cap.ReadLive(source)
	} else if mode == "-pcap" {
		err = cap.ReadPcap(source)
	} else {
		help()
		err = errors.New("Invalid arguments")
	}

	if err != nil {
		fmt.Println(err)
	}
}
