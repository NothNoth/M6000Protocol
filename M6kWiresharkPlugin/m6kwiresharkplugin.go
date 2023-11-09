package main

import (
	"encoding/binary"
	"log"
	"os"

	"m6kparse/common"
	"m6kparse/midi"
	"m6kparse/tcpparser"

	"gitlab.gb/bgirard/wirego/wirego/wirego"
)

var fields []wirego.WiresharkField

// Define here enum identifiers, used to refer to a specific field
const (
	FieldIdDiscoveryMagic wirego.FieldId = 1
	FieldIdFrameSerial    wirego.FieldId = 2
	FieldIdMessagesCount  wirego.FieldId = 3
	FieldIdMessagesNumber wirego.FieldId = 4
	FieldIdFileName       wirego.FieldId = 5
	FieldIdFrameName      wirego.FieldId = 6

	FieldIdProtoVersion wirego.FieldId = 7
	FieldIdBlockSize    wirego.FieldId = 8
	FieldIdMessageType  wirego.FieldId = 9
)

// Since we implement the wirego.WiregoInterface we need some structure to hold it.
type WiregoM6k struct {
	iconIdentified bool
	iconIP         string
	frameIP        string
	parser         *tcpparser.TCPParser
	midi           *midi.MIDI
	log            *log.Logger
}

// Unused (but mandatory)
func main() {}

// Called at golang environment initialization (you should probably not touch this)
func init() {
	var wgo WiregoM6k

	wgo.iconIdentified = false
	//Register to the wirego package
	wirego.Register(&wgo)

	wgo.log = log.New(os.Stdout, "Wirego", 0)
	wgo.midi = midi.New(wgo.log)

	wgo.parser = nil

}

// This function is called when the plugin is loaded.
func (wgo *WiregoM6k) Setup() error {

	//Setup our wireshark custom fields
	fields = append(fields, wirego.WiresharkField{InternalId: FieldIdDiscoveryMagic, Name: "Discovery magic", Filter: "wirego.magic", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeHexadecimal})
	fields = append(fields, wirego.WiresharkField{InternalId: FieldIdFrameSerial, Name: "Serial number", Filter: "wirego.serial", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})
	fields = append(fields, wirego.WiresharkField{InternalId: FieldIdMessagesCount, Name: "Msg count", Filter: "wirego.msgcount", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})
	fields = append(fields, wirego.WiresharkField{InternalId: FieldIdMessagesNumber, Name: "Msg number", Filter: "wirego.msgnum", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})
	fields = append(fields, wirego.WiresharkField{InternalId: FieldIdFileName, Name: "File name", Filter: "wirego.filename", ValueType: wirego.ValueTypeCString, DisplayMode: wirego.DisplayModeNone})
	fields = append(fields, wirego.WiresharkField{InternalId: FieldIdFrameName, Name: "Frame name", Filter: "wirego.framename", ValueType: wirego.ValueTypeCString, DisplayMode: wirego.DisplayModeNone})
	fields = append(fields, wirego.WiresharkField{InternalId: FieldIdProtoVersion, Name: "Protocol version", Filter: "wirego.version", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})
	fields = append(fields, wirego.WiresharkField{InternalId: FieldIdBlockSize, Name: "Block size", Filter: "wirego.bs", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})

	return nil
}

// This function shall return the plugin name
func (wgo *WiregoM6k) GetName() string {
	return "TC M6000"
}

// This function shall return the wireshark filter
func (wgo *WiregoM6k) GetFilter() string {
	return "tcm6000"
}

// GetFields returns the list of fields descriptor that we may eventually return
// when dissecting a packet payload
func (wgo *WiregoM6k) GetFields() []wirego.WiresharkField {
	return fields
}

// GetDissectorFilter returns a wireshark filter that will select which packets
// will be sent to your dissector for parsing.
// Two types of filters can be defined: Integers or Strings
func (wgo *WiregoM6k) GetDissectorFilter() []wirego.DissectorFilter {
	var filters []wirego.DissectorFilter

	filters = append(filters, wirego.DissectorFilter{FilterType: wirego.DissectorFilterTypeInt, Name: "udp.port", ValueInt: 17})
	filters = append(filters, wirego.DissectorFilter{FilterType: wirego.DissectorFilterTypeInt, Name: "tcp.port", ValueInt: 1026})

	return filters
}
func (wgo *WiregoM6k) DissectPacket(src string, dst string, layer string, packet []byte) *wirego.DissectResult {
	if layer == "frame.eth.ethertype.ip.tcp.tcm6000" {
		return wgo.DissectPacketTCP(src, dst, layer, packet)
	} else if layer == "frame.eth.ethertype.ip.udp.tcm6000" {
		return wgo.DissectPacketUDP(src, dst, layer, packet)
	} else {
		var res wirego.DissectResult
		return &res
	}
}

// DissectPacket provides the packet payload to be parsed.
func (wgo *WiregoM6k) DissectPacketUDP(src string, dst string, layer string, packet []byte) *wirego.DissectResult {
	var res wirego.DissectResult

	//This string will appear on the packet being parsed
	res.Protocol = "TC Discovery"

	//Check magic 0x12345678
	if (len(packet) >= 4) && (packet[0] != 0x12 || packet[1] != 0x34 || packet[2] != 0x56 || packet[3] != 0x78) {
		res.Info = "not identified (no magic)"
		return &res
	}
	res.Fields = append(res.Fields, wirego.DissectField{InternalId: FieldIdDiscoveryMagic, Offset: 0, Length: 4})

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
		parseIconToFrame(packet, &res)
	}

	//Frame to icon message
	if dst == wgo.iconIP {
		//If not already known, identify responding frame
		if wgo.frameIP == "" {
			wgo.frameIP = src
		}
		parseFrameToIcon(packet, &res)
	}

	return &res
}

// DissectPacket provides the packet payload to be parsed.
func (wgo *WiregoM6k) DissectPacketTCP(src string, dst string, layer string, packet []byte) *wirego.DissectResult {
	var res wirego.DissectResult

	if wgo.parser == nil {
		wgo.parser = tcpparser.New(wgo.iconIP, wgo.frameIP, wgo.log)
	}
	/*
		if src == wgo.frameIP {
			wgo.parser.ParseBlocks(packet, common.FrameToIcon)
		} else {
			wgo.parser.ParseBlocks(packet, common.IconToFrame)
		}*/

	//This string will appear on the packet being parsed
	res.Protocol = "TC Proto"

	res.Fields = append(res.Fields, wirego.DissectField{InternalId: FieldIdProtoVersion, Offset: 0, Length: 2})
	res.Fields = append(res.Fields, wirego.DissectField{InternalId: FieldIdBlockSize, Offset: 2, Length: 2})

	if (binary.BigEndian.Uint16(packet[:2]) == 0x0002) && (binary.BigEndian.Uint16(packet[2:4]) == uint16(len(packet)-4)) {
		if src == wgo.frameIP {
			res.Info = wgo.midi.Parse(packet[4:], common.FrameToIcon)
		} else {
			res.Info = wgo.midi.Parse(packet[4:], common.IconToFrame)
		}
	} else {
		res.Info = "Split block (unsupported)"
	}
	//fmt.Println(hex.Dump(packet))

	return &res
}

func parseIconToFrame(packet []byte, result *wirego.DissectResult) error {
	result.Info = "Icon probe"

	return nil
}

func parseFrameToIcon(packet []byte, result *wirego.DissectResult) error {
	result.Info = "Mainframe probe response"
	result.Fields = append(result.Fields, wirego.DissectField{InternalId: FieldIdFrameSerial, Offset: 4, Length: 4})

	result.Fields = append(result.Fields, wirego.DissectField{InternalId: FieldIdMessagesCount, Offset: 8, Length: 1})
	result.Fields = append(result.Fields, wirego.DissectField{InternalId: FieldIdMessagesNumber, Offset: 9, Length: 1})

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
	result.Fields = append(result.Fields, wirego.DissectField{InternalId: FieldIdFileName, Offset: idxStart, Length: length})

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
	result.Fields = append(result.Fields, wirego.DissectField{InternalId: FieldIdFrameName, Offset: idxStart, Length: length})

	return nil
}
