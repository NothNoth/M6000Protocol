package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"m6kparse/common"
	"m6kparse/midi"
	"m6kparse/tcpparser"
	"os"

	"gitlab.qb/bgirard/wirego/wirego/wirego"
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
	fields = append(fields, wirego.WiresharkField{WiregoFieldId: FieldIdDiscoveryMagic, Name: "Discovery magic", Filter: "tcm6000.magic", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeHexadecimal})
	fields = append(fields, wirego.WiresharkField{WiregoFieldId: FieldIdFrameSerial, Name: "Serial number", Filter: "tcm6000.serial", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})
	fields = append(fields, wirego.WiresharkField{WiregoFieldId: FieldIdMessagesCount, Name: "Msg count", Filter: "tcm6000.msgcount", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})
	fields = append(fields, wirego.WiresharkField{WiregoFieldId: FieldIdMessagesNumber, Name: "Msg number", Filter: "tcm6000.msgnum", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})
	fields = append(fields, wirego.WiresharkField{WiregoFieldId: FieldIdFileName, Name: "File name", Filter: "tcm6000.filename", ValueType: wirego.ValueTypeCString, DisplayMode: wirego.DisplayModeNone})
	fields = append(fields, wirego.WiresharkField{WiregoFieldId: FieldIdFrameName, Name: "Frame name", Filter: "tcm6000.framename", ValueType: wirego.ValueTypeCString, DisplayMode: wirego.DisplayModeNone})
	fields = append(fields, wirego.WiresharkField{WiregoFieldId: FieldIdProtoVersion, Name: "Protocol version", Filter: "tcm6000.version", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})
	fields = append(fields, wirego.WiresharkField{WiregoFieldId: FieldIdBlockSize, Name: "Block size", Filter: "tcm6000.bs", ValueType: wirego.ValueTypeUInt32, DisplayMode: wirego.DisplayModeDecimal})

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

func (wgo *WiregoM6k) DissectPacket(packetNumber int, src string, dst string, layer string, packet []byte) *wirego.DissectResult {
	if layer == "frame.eth.ethertype.ip.tcp.tcm6000" {
		return wgo.DissectPacketTCP(packetNumber, src, dst, layer, packet)
	} else if layer == "frame.eth.ethertype.ip.udp.tcm6000" {
		return wgo.DissectPacketUDP(packetNumber, src, dst, layer, packet)
	} else {
		var res wirego.DissectResult
		fmt.Println("Unknown layer:" + layer)
		return &res
	}
}

// DissectPacket provides the packet payload to be parsed.
func (wgo *WiregoM6k) DissectPacketTCP(packetNumber int, src string, dst string, layer string, packet []byte) *wirego.DissectResult {
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

	res.Fields = append(res.Fields, wirego.DissectField{WiregoFieldId: FieldIdProtoVersion, Offset: 0, Length: 2})
	res.Fields = append(res.Fields, wirego.DissectField{WiregoFieldId: FieldIdBlockSize, Offset: 2, Length: 2})

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
