package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	tcFrameDetectionMagic = 0x12345678
)

const (
	SYXTYPE_PRESETDATA    = 0x20
	SYXTYPE_RHYTHMDATA    = 0x21
	SYXTYPE_PARAMDATA     = 0x22
	SYXTYPE_BANKREQUEST   = 0x40
	SYXTYPE_PRESETRECALL  = 0x44
	SYXTYPE_PRESETREQUEST = 0x45
	SYXTYPE_RHYTHMREQUEST = 0x46
	SYXTYPE_PARAMREQUEST  = 0x47
)

type MIDIMessage struct {
	data  []byte
	count uint32
}

var index map[string]MIDIMessage

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

type block struct {
	msgIdx  int
	version int
	data    []byte
}

func parseBlock(payload []byte) []block {
	var blockList []block

	offs := 0
	msg := 1
	for {
		if offs+4 > len(payload) {
			fmt.Println("!!! TRUNCATED block")
			return blockList
		}
		version := int(binary.BigEndian.Uint16(payload[offs : offs+2]))
		offs += 2
		size := int(binary.BigEndian.Uint16(payload[offs : offs+2]))
		offs += 2
		if offs+size <= len(payload) {
			blockList = append(blockList, block{msgIdx: msg, version: version, data: payload[offs : offs+size]})
			//Note: on large packets, truncated message is found on the next packet
		} else {
			fmt.Println("!!! TRUNCATED block")
			return blockList
		}
		offs += size
		if offs == len(payload) {
			return blockList
		}
		msg++
	}

}

func addToIndex(b block) MIDIMessage {
	h := sha1.Sum(b.data)
	hs := hex.EncodeToString(h[:])
	blockIdx, found := index[hs]
	if found {
		blockIdx.count++
		index[hs] = blockIdx
		return blockIdx
	} else {
		blockIdx.count = 1
		blockIdx.data = make([]byte, len(b.data))
		copy(blockIdx.data, b.data)
		index[hs] = blockIdx
		return blockIdx
	}
}

func parseFrameToIconTCP(packet gopacket.Packet, ip *layers.IPv4, tcp *layers.TCP) {

	if len(tcp.Payload) == 0 {
		return
	}
	fmt.Println("------------------------------------------------------- Frame to Icon (tcp)")
	fmt.Printf("Payload size: %d (0x%x) bytes\n", len(tcp.Payload), len(tcp.Payload))
	//fmt.Println(hex.Dump(tcp.Payload))

	blocks := parseBlock(tcp.Payload)
	for _, b := range blocks {
		fmt.Println(addToIndex(b))
		//fmt.Printf("Block %d (v: %d)\n", b.msgIdx, b.version)
		//fmt.Println(hex.Dump(b.data))
	}

}

func parseIconToFrameTCP(packet gopacket.Packet, ip *layers.IPv4, tcp *layers.TCP) {

	if len(tcp.Payload) == 0 {
		return
	}
	fmt.Println("------------------------------------------------------- Icon to Frame (tcp)")
	fmt.Printf("Payload size: %d (0x%x) bytes\n", len(tcp.Payload), len(tcp.Payload))
	//fmt.Println(hex.Dump(tcp.Payload))
	blocks := parseBlock(tcp.Payload)
	for _, b := range blocks {
		fmt.Println(addToIndex(b))
		//fmt.Printf("Block %d (v: %d)\n", b.msgIdx, b.version)
		//fmt.Println(hex.Dump(b.data))
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

func messageTypeToString(msgType byte) string {
	switch msgType {
	case SYXTYPE_PRESETDATA:
		return "PresetData"
	case SYXTYPE_RHYTHMDATA:
		return "RythmData"
	case SYXTYPE_PARAMDATA:
		return "ParamData"
	case SYXTYPE_PRESETREQUEST:
		return "PresetRequest"
	case SYXTYPE_RHYTHMREQUEST:
		return "RythmRequest"
	case SYXTYPE_PARAMREQUEST:
		return "ParamRequest"
	case SYXTYPE_BANKREQUEST:
		return "BankRequest"
	case SYXTYPE_PRESETRECALL:
		return "PresetRecall"
	}
	return "Unknown"
}

func (midiMsg MIDIMessage) String() string {
	var str string

	if midiMsg.data[0] == 0xFF {
		return "MIDI Reset message"
	}
	if (midiMsg.data[0] != 0xF0) || (midiMsg.data[len(midiMsg.data)-1] != 0xF7) {
		return "Not a SysEx message:" + hex.Dump(midiMsg.data)
	}
	msg := midiMsg.data[1 : len(midiMsg.data)-1]
	manufacturerID := msg[0:3]
	sysExDeviceID := msg[3]
	modelID := msg[4]
	messageType := msg[5]
	messageData := msg[6:]

	if (manufacturerID[0] != 0x00) || (manufacturerID[1] != 0x20) || (manufacturerID[2] != 0x1F) {
		return "Not a TC Electronic manufacturer:" + hex.Dump(midiMsg.data)
	}
	if modelID != 0x46 {
		return "Not a M6000 device ID:" + hex.Dump(midiMsg.data)
	}
	str += fmt.Sprintf("SysExDeviceID: 0x%02x | MessageType: 0x%02x (%s)\n", sysExDeviceID, messageType, messageTypeToString(messageType))

	if messageType == SYXTYPE_PRESETRECALL {
		engine := msg[6]
		presetMSB := msg[7]
		presetLSB := msg[8]
		preset := (uint16(presetMSB&0x7F) << 7) | uint16(presetLSB&0x7F)
		str += fmt.Sprintf("Recall preset %d on engine %d\n", preset, engine)
	} else if messageType == SYXTYPE_PRESETREQUEST {
		presetMSB := msg[6]
		presetLSB := msg[7]
		preset := (uint16(presetMSB&0x7F) << 7) | uint16(presetLSB&0x7F)
		str += fmt.Sprintf("Request preset %d\n", preset)
	} else if messageType == SYXTYPE_PRESETDATA {

	}
	str += hex.Dump(messageData)
	return str
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: " + os.Args[0] + " <pcap file> <frame IP> <icon IP>")
		fmt.Println("Example: " + os.Args[0] + " boot.pcap 192.168.1.126 192.168.1.125")
		return
	}

	pcapFileName := os.Args[1]
	frameIP := os.Args[2]
	iconIP := os.Args[3]

	index = make(map[string]MIDIMessage)

	h, err := pcap.OpenOffline(pcapFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	packetSource := gopacket.NewPacketSource(h, h.LinkType())
	for packet := range packetSource.Packets() {
		ipLayer := packet.Layer(layers.LayerTypeIPv4)

		if ipLayer != nil {
			var tcp *layers.TCP
			var udp *layers.UDP
			ip, _ := ipLayer.(*layers.IPv4)

			udpLayer := packet.Layer((layers.LayerTypeUDP))
			if udpLayer != nil {
				udp, _ = udpLayer.(*layers.UDP)
			}
			tcpLayer := packet.Layer((layers.LayerTypeTCP))
			if tcpLayer != nil {
				tcp, _ = tcpLayer.(*layers.TCP)
			}

			if (ip.SrcIP.String() == frameIP) && (ip.DstIP.String() == iconIP) {
				if tcp != nil {
					parseFrameToIconTCP(packet, ip, tcp)
				} else if udp != nil {
					parseFrameToIconUDP(packet, ip, udp)
				} else {
					fmt.Println("Frame to Icon (unkown):")
					fmt.Println(packet)
				}
			} else if (ip.SrcIP.String() == iconIP) && (ip.DstIP.String() == frameIP) {
				if tcp != nil {
					parseIconToFrameTCP(packet, ip, tcp)
				} else if udp != nil {
					parseIconToFrameUDP(packet, ip, udp)
				} else {
					fmt.Println("Icon to Frame (unkown):")
					fmt.Println(packet)
				}
			} else if (ip.SrcIP.String() == iconIP) && (ip.DstIP[3] == 255) {
				parseIconToBroadcastUDP(packet, ip, udp)
			} else if ip.SrcIP.String() == frameIP {
				fmt.Println("Frame to unknown (" + ip.DstIP.String() + "):")
				fmt.Println(hex.Dump(ip.Payload))
			} else if ip.SrcIP.String() == iconIP {
				fmt.Println("Icon to unknown (" + ip.DstIP.String() + "):")
				fmt.Println(hex.Dump(ip.Payload))
			}
		}
	}
}
