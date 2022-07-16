package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
)

type MIDIMessage struct {
	data []byte
}

const (
	//0x25 difference between request and response
	SYXTYPE_PRESETDATA    = 0x20
	SYXTYPE_PRESETREQUEST = 0x45 // 5 bytes messages

	SYXTYPE_RHYTHMDATA    = 0x21 // 72 or 124 bytes messages
	SYXTYPE_RHYTHMREQUEST = 0x46 // 9 bytes messages

	SYXTYPE_PARAMDATA    = 0x22
	SYXTYPE_PARAMREQUEST = 0x47 // 6 bytes messages

	SYXTYPE_BANKREQUEST  = 0x40
	SYXTYPE_PRESETRECALL = 0x44

	SYXTYPE_UNKNOWN_23 = 0x23 // variable message len
	SYXTYPE_UNKNOWN_28 = 0x28 // 5 bytes messages
	SYXTYPE_UNKNOWN_29 = 0x29 // 3 or 72 bytes messages
	SYXTYPE_UNKNOWN_2F = 0x2F // variable message len
	SYXTYPE_UNKNOWN_43 = 0x43
	SYXTYPE_UNKNOWN_49 = 0x49 // 3 bytes messages
	SYXTYPE_UNKNOWN_4A = 0x4A // 3 bytes messages
	SYXTYPE_UNKNOWN_4F = 0x4F // 3 bytes messages

	SYXTYPE_UNKNOWN_4E = 0x4E
	SYXTYPE_UNKNOWN_2E = 0x2E
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

	//Message Type dependant print

	/*
		switch messageType {
		case SYXTYPE_PRESETREQUEST: // M-One spec or D-Two
			if len(messageData) != 2 {
				str += "[Format length mismatch]"
			}
			presetMSB := messageData[0]
			presetLSB := messageData[1]
			preset := (uint16(presetMSB&0x7F) << 7) | uint16(presetLSB&0x7F)
			str += fmt.Sprintf("Request preset %d\n", preset)
		case SYXTYPE_PRESETDATA: // D-Two spec
			//zero := messageData[6] // on M One
			presetMSB := messageData[0]
			presetLSB := messageData[1]
			preset := (uint16(presetMSB&0x7F) << 7) | uint16(presetLSB&0x7F)
			str += fmt.Sprintf("Preset Data for %d (TODO)\n", preset) // See D-Two spec page 2
		case SYXTYPE_PARAMREQUEST: // M-One spec or D-Two
			if len(messageData) != 2 {
				str += "[Format length mismatch]"
			}
			engine := messageData[0]
			paramId := messageData[1]
			str += fmt.Sprintf("Param request for engine %d param: %d\n", engine, paramId)
		case SYXTYPE_PARAMDATA: // M-One spec or D-Two
			if len(messageData) != 4 {
				str += "[Format length mismatch]"
			}
			engine := messageData[0]
			paramId := messageData[1]
			valuetMSB := messageData[2]
			valueLSB := messageData[3]
			value := (uint16(valuetMSB&0x7F) << 7) | uint16(valueLSB&0x7F)
			str += fmt.Sprintf("Param data for engine %d param: %d has value %d\n", engine, paramId, value)

		case SYXTYPE_RHYTHMREQUEST:
			if len(messageData) != 0 {
				str += "[Format length mismatch]"
			}
		case SYXTYPE_RHYTHMDATA:
			if len(messageData) != 44 {
				str += "[Format length mismatch] TODO"
			}
		}
	*/

	str += "MessageData:\n"
	str += hex.Dump(messageData)

	h := sha1.Sum(messageData)
	hs := hex.EncodeToString(h[:])
	os.WriteFile(fmt.Sprintf("%02x_%s.bin", messageType, hs), messageData, 0777)
	return str
}
