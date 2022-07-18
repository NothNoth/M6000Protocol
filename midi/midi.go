package midi

import (
	"encoding/hex"
	"fmt"
	"log"
	"m6kparse/common"
)

type MIDI struct {
	logs                      *log.Logger
	truncatedSysExFrameToIcon []byte
	truncatedSysExIconToFrame []byte
}

type MIDIType int

const (
	MIDITypeReset MIDIType = iota
	MIDITypeSysEx MIDIType = iota
)

type MIDIMessage struct {
	data     []byte
	midiType MIDIType
}

func New(logs *log.Logger) *MIDI {
	var m MIDI
	m.logs = logs
	return &m
}

func (m *MIDI) Parse(midiData []byte, dir common.Direction) {
	var msg MIDIMessage

	//Previous message was truncated, continue
	if (dir == common.FrameToIcon) && (len(m.truncatedSysExFrameToIcon) != 0) {
		msg.data = append(m.truncatedSysExFrameToIcon, midiData...)
		m.truncatedSysExFrameToIcon = make([]byte, 0)
	} else if (dir == common.IconToFrame) && (len(m.truncatedSysExIconToFrame) != 0) {
		msg.data = append(m.truncatedSysExIconToFrame, midiData...)
		m.truncatedSysExIconToFrame = make([]byte, 0)
	} else {
		msg.data = make([]byte, len(midiData))
		copy(msg.data, midiData)
	}

	if msg.data[0] == 0xFF {
		msg.midiType = MIDITypeReset
	} else if msg.data[0] == 0xF0 {
		if msg.data[len(msg.data)-1] == 0xF7 {
			msg.midiType = MIDITypeSysEx
		} else {
			//Truncated SysEx
			if dir == common.FrameToIcon {
				m.truncatedSysExFrameToIcon = make([]byte, len(msg.data))
				copy(m.truncatedSysExFrameToIcon, msg.data)
			} else if dir == common.IconToFrame {
				m.truncatedSysExIconToFrame = make([]byte, len(msg.data))
				copy(m.truncatedSysExIconToFrame, msg.data)
			}
		}
	}

	switch msg.midiType {
	case MIDITypeReset:

	}
	m.logs.Println(msg)
}

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

	SYXTYPE_UNKNOWN_4E = 0x4E // variable message len
	SYXTYPE_UNKNOWN_2E = 0x2E // 3 bytes messages
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
	case SYXTYPE_UNKNOWN_2E:
		return "Frame to icon unknown 2E (Licence?)"
	case SYXTYPE_UNKNOWN_2F:
		return "Frame to icon unknown 2F"
	case SYXTYPE_UNKNOWN_43:
		return "Icon to frame unknown 43"
	case SYXTYPE_UNKNOWN_49:
		return "Icon to frame unknown 49"
	case SYXTYPE_UNKNOWN_4A:
		return "Icon to frame unknown 4A"
	case SYXTYPE_UNKNOWN_4E:
		return "Icon to frame unknown 4E (Licence?)"
	case SYXTYPE_UNKNOWN_4F:
		return "Icon to frame unknown 4F"
	}
	return "Unknown"
}

func midiTwoBytesTo14Bits(a byte, b byte) uint16 {
	return ((uint16(a) & 0x7F) << 7) | uint16(b)
}

func midiToBytesToSigned14Bits(a byte, b byte) int16 {
	var value int16

	value = ((int16(a) & 0x7F) << 7) | int16(b)
	if value&0x2000 == 1 { // "sign" bit is set?
		value |= 0x2000 //erase it
		return -value   //return negative value
	}
	return value
}

var split []byte

func (midiMsg MIDIMessage) String() string {
	var str string

	if midiMsg.data[0] == 0xFF {
		return "MIDI Reset message"
	}
	/*
	   //FIXME
	   	if len(split) != 0 {
	   		midiMsg.data = append(split, midiMsg.data...)
	   		split = make([]byte, 0)
	   		str += fmt.Sprintf("reuse %d bytes\n", len(split))
	   	}
	*/
	if (midiMsg.data[0] != 0xF0) || (midiMsg.data[len(midiMsg.data)-1] != 0xF7) {
		split = make([]byte, len(midiMsg.data))
		copy(split, midiMsg.data)
		return "[Error] Not a SysEx message:" + hex.Dump(midiMsg.data)
	}
	msg := midiMsg.data[1 : len(midiMsg.data)-1]
	manufacturerID := msg[0:3] // Should be TC ident 00201F
	sysExDeviceID := msg[3]    //Configurable using the Icon
	modelID := msg[4]          // Should be 0x46 for M6000
	messageType := msg[5]      //More or less stable for TC product range
	messageData := msg[6:]     //Slight variations between product ranges

	//Make sure this is TC
	if (manufacturerID[0] != 0x00) || (manufacturerID[1] != 0x20) || (manufacturerID[2] != 0x1F) {
		return "[Error] Not a TC Electronic manufacturer:" + hex.Dump(midiMsg.data)
	}
	//Make sure it's M6000
	if modelID != 0x46 {
		return "[Error] Not a M6000 device ID:" + hex.Dump(midiMsg.data)
	}
	str += fmt.Sprintf("SysExDeviceID: 0x%02x | MessageType: 0x%02x (%s)\n", sysExDeviceID, messageType, messageTypeToString(messageType))

	str += fmt.Sprintf("MessageData (%d bytes)\n", len(messageData))
	str += hex.Dump(messageData)
	str += "\n"
	//Message Type dependant print
	switch messageType {
	case SYXTYPE_PARAMREQUEST:
		engine := messageData[0]
		paramId := messageData[1]
		unkA := messageData[2]
		unkB := messageData[3]
		count := midiTwoBytesTo14Bits(messageData[4], messageData[5])
		str += fmt.Sprintf("[Parsed] Param request for engine %d param: %d / Count: %d / Response size: %d [unkA/B : x%02x x%02x]\n", engine, paramId, count, count*2+4, unkA, unkB)

	case SYXTYPE_PARAMDATA:
		engine := messageData[0]
		paramId := messageData[1]
		unkA := messageData[2]
		unkB := messageData[3]
		str += fmt.Sprintf("[Parsed] Param data for engine %d param: %d  [unkA/B : x%02x x%02x]\nValues:\n", engine, paramId, unkA, unkB)

		/*
			for offs := 4; offs < len(messageData); offs += 2 {
				value := midiToBytesToSigned14Bits(messageData[offs], messageData[offs+1])
				str += fmt.Sprintf("0x%04x %+d (%c)\n", value, value, value)
			}
		*/
	}

	/*
		h := sha1.Sum(messageData)
		hs := hex.EncodeToString(h[:])
		os.WriteFile(fmt.Sprintf("%02x_%s.bin", messageType, hs), messageData, 0777)
	*/

	return str
}
