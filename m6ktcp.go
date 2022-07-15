package main

import (
	"encoding/binary"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var truncatedIconToFrame []byte
var truncatedFrameToIcon []byte

func parseFrameToIconTCP(packet gopacket.Packet, ip *layers.IPv4, tcp *layers.TCP) {
	var blocks []midiOverIPMessage
	if len(tcp.Payload) == 0 {
		return
	}
	fmt.Println("------------------------------------------------------- Frame to Icon (tcp)")
	fmt.Printf("Payload size: %d (0x%x) bytes\n", len(tcp.Payload), len(tcp.Payload))

	blocks, truncatedFrameToIcon = parseBlock(tcp.Payload)
	for _, b := range blocks {
		fmt.Println(addToIndex(b))
	}

}

func parseIconToFrameTCP(packet gopacket.Packet, ip *layers.IPv4, tcp *layers.TCP) {
	var blocks []midiOverIPMessage
	if len(tcp.Payload) == 0 {
		return
	}
	fmt.Println("------------------------------------------------------- Icon to Frame (tcp)")
	fmt.Printf("Payload size: %d (0x%x) bytes\n", len(tcp.Payload), len(tcp.Payload))
	blocks, truncatedIconToFrame = parseBlock(tcp.Payload)
	for _, b := range blocks {
		fmt.Println(addToIndex(b))
	}
}

type midiOverIPMessage struct {
	msgIdx  int
	version int
	data    []byte
}

func parseBlock(payload []byte) ([]midiOverIPMessage, []byte) {
	var blockList []midiOverIPMessage

	offs := 0
	msg := 1
	for {
		if offs+4 > len(payload) {
			fmt.Println("!!! TRUNCATED block")
			return blockList, payload[offs:]
		}
		version := int(binary.BigEndian.Uint16(payload[offs : offs+2]))
		offs += 2
		size := int(binary.BigEndian.Uint16(payload[offs : offs+2]))
		offs += 2
		if offs+size <= len(payload) {
			blockList = append(blockList, midiOverIPMessage{msgIdx: msg, version: version, data: payload[offs : offs+size]})
			//Note: on large packets, truncated message is found on the next packet
		} else {
			fmt.Println("!!! TRUNCATED block")
			return blockList, payload[offs:]
		}
		offs += size
		if offs == len(payload) {
			return blockList, []byte{}
		}
		msg++
	}

}
