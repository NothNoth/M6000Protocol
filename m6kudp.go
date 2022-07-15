package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func parseFrameToIconUDP(packet gopacket.Packet, ip *layers.IPv4, udp *layers.UDP) {
	magic := binary.BigEndian.Uint32(udp.Payload[0:4])
	if magic != tcFrameDetectionMagic {
		//Otehr UDP packets are Timeframes sent from Frame to Icon.
		return
	}
	fmt.Println("------------------------------------------------------- Frame to icon (udp)")
	fmt.Printf("Payload size: %d (0x%x) bytes\n", len(udp.Payload), len(udp.Payload))
	fmt.Println(hex.Dump(udp.Payload))

	frameSerial := binary.BigEndian.Uint32(udp.Payload[4:8])
	totalMsg := udp.Payload[8] //not sure
	unknownA := udp.Payload[0x9:0x10]
	currentMsg := udp.Payload[0x13] //not sure
	fileName := string(udp.Payload[0x14:0x27])
	deviceName := string(udp.Payload[0x54:0x67])
	fmt.Println("  Serial: ", frameSerial)
	fmt.Printf("  Message: %d/%d\n", currentMsg+1, totalMsg)
	fmt.Println(hex.Dump(unknownA))
	fmt.Println("  Filename: " + fileName)
	fmt.Println("  DeviceName: " + deviceName)
}

func parseIconToFrameUDP(packet gopacket.Packet, ip *layers.IPv4, udp *layers.UDP) {
	magic := binary.BigEndian.Uint32(udp.Payload[0:4])
	if magic != tcFrameDetectionMagic {
		return
	}
	fmt.Println("------------------------------------------------------- Icon to Frame (udp)")

	if magic == tcFrameDetectionMagic {
		fmt.Println("Icon response to frame")
		fmt.Println(hex.Dump(udp.Payload))
		command := string(udp.Payload[4:16])
		fmt.Println("  Icon command " + command)
	}
}

func parseIconToBroadcastUDP(packet gopacket.Packet, ip *layers.IPv4, udp *layers.UDP) {

	if udp.DstPort == 137 || udp.DstPort == 138 {
		//Ignore all netbios stuff
		return
	}
	fmt.Println("-------------------------------------------------------")

	/*
		fmt.Println(packet)
		fmt.Printf("Payload size: %d (0x%x) bytes\n", len(udp.Payload), len(udp.Payload))
		fmt.Println(hex.Dump(udp.Payload))
	*/
	magic := binary.BigEndian.Uint32(udp.Payload[0:4])
	if magic != tcFrameDetectionMagic {
		fmt.Printf("--> Invalid TC Magic: 0x%08x\n", magic)
		return
	}
	name := string(udp.Payload[4:16])
	fmt.Println("Icon detect message:")
	fmt.Printf("  Magic %08x\n", magic)
	fmt.Println("  Name: " + name)
	fmt.Println(hex.Dump(udp.Payload[16:]))
}
