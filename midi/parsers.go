package midi

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

func (midiMsg MIDIMessage) parseParamRequest(messageData []byte) string {
	engine := messageData[0]
	paramId := messageData[1]
	unkA := messageData[2]
	unkB := messageData[3]
	count := midiTwoBytesTo14Bits(messageData[4], messageData[5])
	return fmt.Sprintf("[Parsed] Param request for engine %d %d parameters starting from param: %d/ Response size: %d [unkA/B : x%02x x%02x]\n", engine, count, paramId, count*2+4, unkA, unkB)
}

func (midiMsg MIDIMessage) parseParamData(messageData []byte) string {
	engine := messageData[0]
	paramId := messageData[1]

	topBit := (messageData[2] << 7)    // 14 bits value
	unknown := messageData[3] | topBit // 14 bits value
	str := fmt.Sprintf("[Parsed] Param data for engine %d param: %d  [unk: x%02x]\nValues:\n", engine, paramId, unknown)

	for offs := 4; offs < len(messageData); offs += 2 {
		value := midiTwoBytesTo8Bits(messageData[offs], messageData[offs+1])
		if strconv.IsPrint(rune(value)) {
			str += fmt.Sprintf("[%02x %02x] 0x%04x %+d (%c)\n", messageData[offs], messageData[offs+1], value, value, value)
		} else {
			str += fmt.Sprintf("[%02x %02x] 0x%04x %+d ( )\n", messageData[offs], messageData[offs+1], value, value)
		}
	}

	return str
}

func (midiMsg MIDIMessage) parsePresetRequest(messageData []byte) string {
	//Two first bytes are preset number
	presetNumber := uint16(messageData[1])<<8 | uint16(messageData[0])
	return fmt.Sprintf("Preset request for preset %d\n", presetNumber)
}

func (midiMsg MIDIMessage) parsePresetData(messageData []byte) string {
	offset := 0

	if len(messageData) < 3 {
		return "Truncated preset data"
	}

	//Two first bytes are preset number
	presetNumber := uint16(messageData[1])<<8 | uint16(messageData[0])
	offset += 2

	//Unknown byte (0x00)
	unkA := messageData[offset]
	offset++

	rnd := rand.Intn(9999)
	tmpFileName := fmt.Sprintf("presetdata-%d.dat", rnd)
	os.WriteFile(tmpFileName, messageData, 0755)

	f, err := os.Create(tmpFileName + "a")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	for i := offset; i < len(messageData); i += 2 {
		b := ((messageData[i]) << 4) | messageData[i+1]
		f.Write([]byte{b})
	}
	f.Close()

	return fmt.Sprintf("Preset %d dump to %s (unkA:%d)\n", presetNumber, tmpFileName, unkA)
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
		return "Code validation response 'Code length is invalid'" // if not 20+1+8
	case 3:
		return "Code validation response 'Code identifier is invalid'" // - replaced
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
