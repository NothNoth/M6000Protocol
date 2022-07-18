package udpparser

import (
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type UDPParser struct {
	iconIP  string
	frameIP string
	logs    *log.Logger
}

const (
	tcFrameDetectionMagic = 0x12345678
)

func New(iconIP string, frameIP string, logs *log.Logger) *UDPParser {
	var p UDPParser

	p.logs = logs
	p.iconIP = iconIP
	p.frameIP = frameIP

	return &p
}

func (p *UDPParser) Parse(packet gopacket.Packet, ip *layers.IPv4, udp *layers.UDP) {

	if (ip.SrcIP.String() == p.frameIP) && (ip.DstIP.String() == p.iconIP) {
		p.parseFrameToIconUDP(packet, ip, udp)
		return
	}

	if (ip.SrcIP.String() == p.iconIP) && (ip.DstIP.String() == p.frameIP) {
		p.parseIconToFrameUDP(packet, ip, udp)
		return
	}

	if (ip.SrcIP.String() == p.iconIP) && (ip.DstIP[3] == 255) {
		p.parseIconToBroadcastUDP(packet, ip, udp)
		return
	}

	p.logs.Println("[UDP] Unknown traffic:")
	p.logs.Println(hex.Dump(packet.Data()))
}

func (p *UDPParser) parseFrameToIconUDP(packet gopacket.Packet, ip *layers.IPv4, udp *layers.UDP) {
	magic := binary.BigEndian.Uint32(udp.Payload[0:4])
	if magic != tcFrameDetectionMagic {
		//Otehr UDP packets are Timeframes sent from Frame to Icon.
		return
	}
	p.logs.Println("------------------------------------------------------- Frame to icon (udp)")
	p.logs.Printf("Payload size: %d (0x%x) bytes\n", len(udp.Payload), len(udp.Payload))
	p.logs.Println(hex.Dump(udp.Payload))

	frameSerial := binary.BigEndian.Uint32(udp.Payload[4:8])
	totalMsg := udp.Payload[8] //not sure
	unknownA := udp.Payload[0x9:0x10]
	currentMsg := udp.Payload[0x13] //not sure
	fileName := string(udp.Payload[0x14:0x27])
	deviceName := string(udp.Payload[0x54:0x67])
	p.logs.Println("  Serial: ", frameSerial)
	p.logs.Printf("  Message: %d/%d\n", currentMsg+1, totalMsg)
	p.logs.Println(hex.Dump(unknownA))
	p.logs.Println("  Filename: " + fileName)
	p.logs.Println("  DeviceName: " + deviceName)
}

func (p *UDPParser) parseIconToFrameUDP(packet gopacket.Packet, ip *layers.IPv4, udp *layers.UDP) {
	magic := binary.BigEndian.Uint32(udp.Payload[0:4])
	if magic != tcFrameDetectionMagic {
		return
	}
	p.logs.Println("------------------------------------------------------- Icon to Frame (udp)")

	if magic == tcFrameDetectionMagic {
		p.logs.Println("Icon response to frame")
		p.logs.Println(hex.Dump(udp.Payload))
		command := string(udp.Payload[4:16])
		p.logs.Println("  Icon command " + command)
	}
}

func (p *UDPParser) parseIconToBroadcastUDP(packet gopacket.Packet, ip *layers.IPv4, udp *layers.UDP) {

	if udp.DstPort == 137 || udp.DstPort == 138 {
		//Ignore all netbios stuff
		return
	}
	p.logs.Println("-------------------------------------------------------")

	/*
		p.logs.Println(packet)
		p.logs.Printf("Payload size: %d (0x%x) bytes\n", len(udp.Payload), len(udp.Payload))
		p.logs.Println(hex.Dump(udp.Payload))
	*/
	magic := binary.BigEndian.Uint32(udp.Payload[0:4])
	if magic != tcFrameDetectionMagic {
		p.logs.Printf("--> Invalid TC Magic: 0x%08x\n", magic)
		return
	}
	name := string(udp.Payload[4:16])
	p.logs.Println("Icon detect message:")
	p.logs.Printf("  Magic %08x\n", magic)
	p.logs.Println("  Name: " + name)
	p.logs.Println(hex.Dump(udp.Payload[16:]))
}
