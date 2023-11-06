package midi

import (
	"encoding/hex"
	"fmt"
	"log"
	"m6kparse/common"
	"os"
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

	f, _ := os.Create("midi.log")
	mlogs := log.New(f, "MIDI", log.Lshortfile)

	m.logs = mlogs
	return &m
}

func (m *MIDI) Parse(midiData []byte, dir common.Direction) {
	var msg MIDIMessage
	m.logs.Println("*********** " + dir.String() + " **********")
	defer m.logs.Println("----------------------------------")
	m.logs.Println("-> MIDI parsing", len(midiData), "bytes of data")

	//Previous message was truncated, reuse and continue
	if (dir == common.FrameToIcon) && (len(m.truncatedSysExFrameToIcon) != 0) {
		m.logs.Printf("-> Reusing %d bytes from previously truncated SysEx message\n", len(m.truncatedSysExFrameToIcon))
		merged := append(m.truncatedSysExFrameToIcon, midiData...)
		msg.data = make([]byte, len(merged))
		copy(msg.data, merged)
		m.truncatedSysExFrameToIcon = nil
		//m.logs.Println("-> SysEx is now:\n ", hex.Dump(msg.data))
	} else if (dir == common.IconToFrame) && (len(m.truncatedSysExIconToFrame) != 0) {
		m.logs.Printf("-> Reusing %d bytes from previously truncated SysEx message\n", len(m.truncatedSysExIconToFrame))
		merged := append(m.truncatedSysExIconToFrame, midiData...)
		msg.data = make([]byte, len(merged))
		copy(msg.data, merged)
		m.truncatedSysExIconToFrame = nil
		//m.logs.Println("-> SysEx is now:\n ", hex.Dump(msg.data))
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
			//m.logs.Printf("-> Saving %d bytes of truncated SysEx message for next time\n", len(msg.data))

			//Truncated SysEx, save for later
			if dir == common.FrameToIcon {
				m.truncatedSysExFrameToIcon = make([]byte, len(msg.data))
				copy(m.truncatedSysExFrameToIcon, msg.data)
			} else if dir == common.IconToFrame {
				m.truncatedSysExIconToFrame = make([]byte, len(msg.data))
				copy(m.truncatedSysExIconToFrame, msg.data)
			}
			return
		}
	} else {
		m.logs.Println("[WARN] Totally unknown message:" + hex.Dump(msg.data))
		return
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
		return "Frame to icon code command (Licence?)"
	case SYXTYPE_UNKNOWN_2F:
		return "Frame to icon unknown 2F"
	case SYXTYPE_UNKNOWN_43:
		return "Icon to frame unknown 43"
	case SYXTYPE_UNKNOWN_49:
		return "Icon to frame unknown 49"
	case SYXTYPE_UNKNOWN_4A:
		return "Icon to frame unknown 4A"
	case SYXTYPE_CODECMD_RESPONSE:
		return "Icon to frame code command reply (Licence?)"
	case SYXTYPE_UNKNOWN_4F:
		return "Icon to frame unknown 4F"
	case SYXTYPE_PRESETCMD_4C:
		return "Icon to frame possible preset command"
	case SYXTYPE_MEDIACMD_4D:
		return "Icon to frame possible media command"
	}
	return "Unknown"
}

func midiTwoBytesTo14Bits(a byte, b byte) uint16 {
	return ((uint16(a) & 0x7F) << 7) | uint16(b)
}

func midiTwoBytesTo8Bits(a byte, b byte) uint8 {
	return a<<7 | b
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

func (midiMsg MIDIMessage) String() string {
	var str string

	if midiMsg.data[0] == 0xFF {
		return "MIDI Reset message"
	}

	if (midiMsg.data[0] != 0xF0) || (midiMsg.data[len(midiMsg.data)-1] != 0xF7) {
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
		str += midiMsg.parseParamRequest(messageData)
	case SYXTYPE_PARAMDATA:
		str += midiMsg.parseParamData(messageData)
	case SYXTYPE_CODECMD:
		str += midiMsg.parseLicenceCode(messageData)
	case SYXTYPE_CODECMD_RESPONSE:
		str += midiMsg.parseLicenceCodeResponse(messageData)
	case SYXTYPE_PRESETREQUEST:
		str += midiMsg.parsePresetRequest(messageData)
	case SYXTYPE_PRESETDATA:
		str += midiMsg.parsePresetData(messageData)
	default:
		str += midiMsg.parseUnknown(messageData)
	}

	/*
		h := sha1.Sum(messageData)
		hs := hex.EncodeToString(h[:])
		os.WriteFile(fmt.Sprintf("%02x_%s.bin", messageType, hs), messageData, 0777)
	*/

	return str
}
