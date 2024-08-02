package main

import (
	"github.com/quarkslab/wirego/wirego/wirego"
)

// DissectPacket provides the packet payload to be parsed.
func (wgo *WiregoM6k) DissectPacketUDP(packetNumber int, src string, dst string, layer string, packet []byte) *wirego.DissectResult {
	var res wirego.DissectResult

	//This string will appear on the packet being parsed
	res.Protocol = "TC Discovery"

	//Check magic 0x12345678
	if (len(packet) >= 4) && (packet[0] != 0x12 || packet[1] != 0x34 || packet[2] != 0x56 || packet[3] != 0x78) {
		res.Info = "not identified (no magic)"
		return &res
	}
	res.Fields = append(res.Fields, wirego.DissectField{WiregoFieldId: FieldIdDiscoveryMagic, Offset: 0, Length: 4})

	//Identify peers (use the icon probe)
	if !wgo.iconIdentified && len(packet) > 10 && (string(packet[4:10]) == "TCIcon") {
		wgo.iconIdentified = true
		wgo.iconIP = src
		wgo.frameIP = ""
	}
	if !wgo.iconIdentified {
		res.Info = "icon not identified"
		return &res
	}

	//Icon to frame message
	if src == wgo.iconIP {
		parseDiscoveryIconToFrame(packet, &res)
	}

	//Frame to icon message
	if dst == wgo.iconIP {
		//If not already known, identify responding frame
		if wgo.frameIP == "" {
			wgo.frameIP = src
		}
		parseDiscoveryFrameToIcon(packet, &res)
	}

	return &res
}

func parseDiscoveryIconToFrame(packet []byte, result *wirego.DissectResult) error {
	result.Info = "Icon probe"

	return nil
}

func parseDiscoveryFrameToIcon(packet []byte, result *wirego.DissectResult) error {
	result.Info = "Mainframe probe response"
	result.Fields = append(result.Fields, wirego.DissectField{WiregoFieldId: FieldIdFrameSerial, Offset: 4, Length: 4})

	result.Fields = append(result.Fields, wirego.DissectField{WiregoFieldId: FieldIdMessagesCount, Offset: 8, Length: 1})
	result.Fields = append(result.Fields, wirego.DissectField{WiregoFieldId: FieldIdMessagesNumber, Offset: 9, Length: 1})

	//File name -> look for trailing \0
	idxStart := 0x14
	length := 0
	for {
		if idxStart+length >= len(packet) {
			break
		}
		if packet[idxStart+length] == 0x00 {
			break
		}
		length++
	}
	result.Fields = append(result.Fields, wirego.DissectField{WiregoFieldId: FieldIdFileName, Offset: idxStart, Length: length})

	//Frame name -> look for trailing \0
	idxStart = 0x54
	length = 0
	for {
		if idxStart+length >= len(packet) {
			break
		}
		if packet[idxStart+length] == 0x00 {
			break
		}
		length++
	}
	result.Fields = append(result.Fields, wirego.DissectField{WiregoFieldId: FieldIdFrameName, Offset: idxStart, Length: length})

	return nil
}
