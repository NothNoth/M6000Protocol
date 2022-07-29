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

	m.logs.Println("-> MIDI parsing", len(midiData), "bytes of data")

	//Previous message was truncated, reuse and continue
	if (dir == common.FrameToIcon) && (len(m.truncatedSysExFrameToIcon) != 0) {
		m.logs.Printf("-> Reusing %d bytes from previously truncated SysEx message\n", len(m.truncatedSysExFrameToIcon))
		merged := append(m.truncatedSysExFrameToIcon, midiData...)
		msg.data = make([]byte, len(merged))
		copy(msg.data, merged)
		m.truncatedSysExFrameToIcon = nil
		m.logs.Println("-> SysEx is now:\n ", hex.Dump(msg.data))
	} else if (dir == common.IconToFrame) && (len(m.truncatedSysExIconToFrame) != 0) {
		m.logs.Printf("-> Reusing %d bytes from previously truncated SysEx message\n", len(m.truncatedSysExIconToFrame))
		merged := append(m.truncatedSysExIconToFrame, midiData...)
		msg.data = make([]byte, len(merged))
		copy(msg.data, merged)
		m.truncatedSysExIconToFrame = nil
		m.logs.Println("-> SysEx is now:\n ", hex.Dump(msg.data))
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
			m.logs.Printf("-> Saving %d bytes of truncated SysEx message for next time\n", len(msg.data))

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
	case SYXTYPE_CODECMD:
		str += midiMsg.parseLicenceCode(messageData)
	case SYXTYPE_CODECMD_RESPONSE:
		str += midiMsg.parseLicenceCodeResponse(messageData)
	}

	/*
		h := sha1.Sum(messageData)
		hs := hex.EncodeToString(h[:])
		os.WriteFile(fmt.Sprintf("%02x_%s.bin", messageType, hs), messageData, 0777)
	*/

	return str
}

func (midiMsg MIDIMessage) parseLicenceCode(msg []byte) string {

	//unk1 := msg[0]
	//fixed7F := msg[1]

	var code string

	offs := 2
	for {
		var ascii byte
		if offs+1 >= len(msg) {
			break
		}

		A := msg[offs]
		B := msg[offs+1]

		ascii = (A << 4) | (B & 0xF) // strings are encoded into 2 bytes per character (1 nibble per byte)
		code += string([]byte{ascii})
		offs += 2
	}

	return "Send code " + code + " to frame for validation"
}

func (midiMsg MIDIMessage) parseLicenceCodeResponse(msg []byte) string {
	if len(msg) != 3 {
		return "Unexpected Code response length"
	}

	result := msg[2]

	switch result {
	case 0:
		return "Code validation response 'Success'"

		/*
			1 = 00713658 + 1*4 = 0071365C => 70FB30 An unspecified error occured
			2 = 00713658 + 2*4 = 00713660 => 71363C Code length is invalid
			3 = 00713658 + 3*4 = 00713664 => 713620 Code identifier is invalid
			4 = 00713658 + 4*4 = 00713668 => 713608 Code level is invalid
			5 = 00713658 + 5*4 = 0071366C => 7135EC Code checksum is invalid
			6 = 00713658 + 6*4 = 00713670 => 7135D0 Device EEPROM is invalid
		*/

	case 1: //Any invalid error code defaults to this
		return "Code validation response 'An unspecified error occured'"
	case 2:
		return "Code validation response 'Code length is invalid'"
	case 3:
		return "Code validation response 'Code identifier is invalid'"
	case 4:
		return "Code validation response 'Code level is invalid'"
	case 5:
		return "Code validation response 'Code checksum is invalid'"
	case 6:
		return "Code validation response 'Device EEPROM is invalid'"
	default:
		return fmt.Sprintf("Code validation response unknown 0x%02x", result)
	}
}
