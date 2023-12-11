package m6000parser

import (
	"fmt"
	"os"
	"strings"
)

const (
	//0x25 difference between request and response
	SYXTYPE_PRESETDATA    = 0x20
	SYXTYPE_PRESETREQUEST = 0x45 // 5 bytes messages

	SYXTYPE_RHYTHMDATA    = 0x21 // 72 or 124 bytes messages
	SYXTYPE_RHYTHMREQUEST = 0x46 // 9 bytes messages

	SYXTYPE_PARAMDATA    = 0x22
	SYXTYPE_PARAMREQUEST = 0x47

	SYXTYPE_BANKREQUEST  = 0x40
	SYXTYPE_PRESETRECALL = 0x44

	SYXTYPE_UNKNOWN_23 = 0x23 // variable message len
	SYXTYPE_UNKNOWN_28 = 0x28 // 5 bytes messages
	SYXTYPE_UNKNOWN_29 = 0x29 // 3 or 72 bytes messages
	SYXTYPE_UNKNOWN_2F = 0x2F // variable message len
	SYXTYPE_UNKNOWN_43 = 0x43 // 5 bytes messages
	SYXTYPE_UNKNOWN_49 = 0x49 // 3 bytes messages
	SYXTYPE_UNKNOWN_4A = 0x4A // 3 bytes messages
	SYXTYPE_UNKNOWN_4F = 0x4F // 3 bytes messages

	SYXTYPE_CODECMD          = 0x4E // variable message len
	SYXTYPE_CODECMD_RESPONSE = 0x2E // 3 bytes messages

	SYXTYPE_PRESETCMD_4C = 0x4C // possible preset command from icon
	SYXTYPE_MEDIACMD_4D  = 0x4D // possible media command from icon

)

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
	case SYXTYPE_UNKNOWN_23:
		return "Frame to icon unknown 23"
	case SYXTYPE_UNKNOWN_28:
		return "Frame to icon unknown 28"
	case SYXTYPE_UNKNOWN_29:
		return "Frame to icon unknown 29"
	case SYXTYPE_CODECMD:
		return "Licence submit)"
	case SYXTYPE_UNKNOWN_2F:
		return "Frame to icon unknown 2F"
	case SYXTYPE_UNKNOWN_43:
		return "Icon to frame unknown 43"
	case SYXTYPE_UNKNOWN_49:
		return "Icon to frame unknown 49"
	case SYXTYPE_UNKNOWN_4A:
		return "Icon to frame unknown 4A"
	case SYXTYPE_CODECMD_RESPONSE:
		return "Licence submit response"
	case SYXTYPE_UNKNOWN_4F:
		return "Icon to frame unknown 4F"
	case SYXTYPE_PRESETCMD_4C:
		return "Icon to frame possible preset command"
	case SYXTYPE_MEDIACMD_4D:
		return "Icon to frame possible media command"
	}
	return fmt.Sprintf("Unk %02x", msgType)
}

// SysEx reassembly
func (m6p *M6000Parser) parseBlock(b blockData) string {
	//m6p.logs.Printf("Parsing block of size %d\n", len(b.data))

	if len(m6p.partialSysex) != 0 {
		b.data = append(m6p.partialSysex, b.data...)
		m6p.partialSysex = []byte{}
		//m6p.logs.Printf("Concat with previous sysex buffer. Current size %d\n", len(b.data))
	}
	if len(b.data) < 3 {
		return "MIDI too short"
	}

	if b.data[0] == 0xF0 && b.data[len(b.data)-1] != 0xF7 {
		m6p.partialSysex = b.data
		m6p.logs.Printf("Incomplete sysex, keep for later %d:\n", len(b.data))
		//m6p.logs.Printf(hex.Dump(m6p.partialSysex))
		return "MIDI Sysex partial"
	}
	//m6p.logs.Printf("Complete sysex %d:\n", len(b.data))
	//m6p.logs.Printf(hex.Dump(m6p.partialSysex))

	//MIDI reset
	if len(b.data) == 3 && b.data[0] == 0xFF && b.data[1] == 0x00 && b.data[2] == 0x00 {
		return "MIDI Reset"
	}

	//MIDI Sysex
	if b.data[0] == 0xF0 && b.data[len(b.data)-1] == 0xF7 {
		return m6p.parseMIDISysex(b.data)
	} else {
		return "MIDI Unknown"
	}
}

var dumpIdx int = 0

func (m6p *M6000Parser) parseMIDISysex(midiMessage []byte) string {

	/*
	 Byte 0 : F0
	 Byte 1-2-3 : Midi identifier for TC Electronic : 00 20 1f
	 Byte 4 : Device Sysex ID (usually 0)
	 Byte 5 : Model id for M6000 0x46
	 Byte 6 : Command
	 ...
	 Byte x : F7

	*/

	command := midiMessage[6]
	payload := midiMessage[7 : len(midiMessage)-1]
	os.WriteFile(fmt.Sprintf("Sysex-%d-%s.sysex", dumpIdx, strings.ReplaceAll(messageTypeToString(command), " ", "-")), payload, 0755)
	dumpIdx++
	p, found := m6p.cmdParsers[command]
	if found {
		return p.Parse(payload)
	}

	return fmt.Sprintf("[%s]", messageTypeToString(command))
}

func midiTwoBytesTo14Bits(a byte, b byte) uint16 {
	return ((uint16(a) & 0x7F) << 7) | uint16(b)
}

func midiTwoBytesTo8Bits(a byte, b byte) uint8 {
	return byte(uint16(a)<<7 | uint16(b))
}
