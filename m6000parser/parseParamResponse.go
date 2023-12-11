package m6000parser

import (
	"fmt"
)

type ParamResponse struct {
}

func (cmd *ParamResponse) Parse(payload []byte) string {
	engine := payload[0]
	paramId := payload[1]

	//topBit := byte(uint16(payload[2]) << 7) // 14 bits value
	//unknown := payload[3] | topBit          // 14 bits value
	str := fmt.Sprintf("Param data (Eng %d - Param: %d)", engine, paramId)
	/*
		for offs := 4; offs < len(payload); offs += 2 {
			value := midiTwoBytesTo8Bits(payload[offs], payload[offs+1])
			if strconv.IsPrint(rune(value)) {
				str += fmt.Sprintf("[%02x %02x] 0x%04x %+d (%c) -", payload[offs], payload[offs+1], value, value, value)
			} else {
				str += fmt.Sprintf("[%02x %02x] 0x%04x %+d ( ) -", payload[offs], payload[offs+1], value, value)
			}
		}*/

	return str
}
