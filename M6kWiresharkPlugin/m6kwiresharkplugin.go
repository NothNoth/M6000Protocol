package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"m6kparse/common"
	"m6kparse/m6000parser"
	"os"
	"strings"

	"github.com/quarkslab/wirego/wirego_remote/go/wirego"
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

	iconToFrameParser *m6000parser.M6000Parser
	frameToIconParser *m6000parser.M6000Parser
	log               *log.Logger
}

// Unused (but mandatory)
func main() {
	var wgo WiregoM6k
	wgo.iconIdentified = false
	wgo.log = log.New(os.Stdout, "Wirego> ", 0)
	wgo.log.Println("m6000 ready")
	wgo.iconToFrameParser = m6000parser.New(wgo.log, common.IconToFrame)
	wgo.frameToIconParser = m6000parser.New(wgo.log, common.FrameToIcon)

	wg, err := wirego.New("ipc:///tmp/wirego0", false, wgo)
	if err != nil {
		fmt.Println(err)
		return
	}
	wg.ResultsCacheEnable(false)

	wg.Listen()
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

// GetDetectionFilters returns a wireshark filter that will select which packets
// will be sent to your dissector for parsing.
// Two types of filters can be defined: Integers or Strings
func (wgo *WiregoM6k) GetDetectionFilters() []wirego.DetectionFilter {
	var filters []wirego.DetectionFilter

	filters = append(filters, wirego.DetectionFilter{FilterType: wirego.DetectionFilterTypeInt, Name: "udp.port", ValueInt: 17})
	filters = append(filters, wirego.DetectionFilter{FilterType: wirego.DetectionFilterTypeInt, Name: "tcp.port", ValueInt: 1026})

	return filters
}

func (wgo *WiregoM6k) GetDetectionHeuristicsParents() []string {
	return []string{}
}

func (wgo *WiregoM6k) DetectionHeuristic(packetNumber int, src string, dst string, layer string, packet []byte) bool {
	return false
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

	//This string will appear on the packet being parsed
	res.Protocol = "TC Proto"

	if !wgo.iconIdentified {
		wgo.iconIdentified = true

		if src == "192.168.1.249" {
			wgo.frameIP = src
			wgo.iconIP = dst
		} else if dst == "192.168.1.249" {
			wgo.frameIP = dst
			wgo.iconIP = src
		} else {
			wgo.frameIP = src
			wgo.iconIP = dst
		}
		//res.Info = "Icon not identified, cannot parse"
		//return &res
	}

	//	res.Fields = append(res.Fields, wirego.DissectField{WiregoFieldId: FieldIdProtoVersion, Offset: 0, Length: 2})
	//	res.Fields = append(res.Fields, wirego.DissectField{WiregoFieldId: FieldIdBlockSize, Offset: 2, Length: 2})

	//	fmt.Println(hex.Dump(packet))
	var parserResult m6000parser.Result

	if src == wgo.iconIP {
		parserResult = wgo.iconToFrameParser.PushPacket(packetNumber, packet)
	} else {
		parserResult = wgo.frameToIconParser.PushPacket(packetNumber, packet)
	}

	var tmp []byte
	fmt.Println("A")
	for i := 0; i < len(packet)-1; i += 2 {
		a := packet[i]
		b := packet[i+1]
		tmp = append(tmp, (a<<4)|(b&0x0F))
	}
	fmt.Println(hex.Dump(tmp))
	fmt.Println("B")
	for i := 1; i < len(packet)-1; i += 2 {
		a := packet[i]
		b := packet[i+1]
		tmp = append(tmp, (a<<4)|(b&0x0F))
	}
	fmt.Println(hex.Dump(tmp))

	aggregate := strings.Join(parserResult.Description, "|")
	switch parserResult.Status {
	case m6000parser.StatusPacketInvalid:
		res.Info = "[Block seems corrupted]" + aggregate
	case m6000parser.StatusPacketSplit:
		res.Info = "[Block is split]" + aggregate
	case m6000parser.StatusPacketSplitFinal:
		res.Info = "[Split End] " + aggregate
	case m6000parser.StatusPacketFull:
		res.Info = aggregate
	default:
		res.Info = "Invalid block status"
	}
	return &res
}
