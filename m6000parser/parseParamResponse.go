package m6000parser

import (
	"encoding/hex"
	"fmt"
)

type ParamResponse struct {
}

func (cmd *ParamResponse) Parse(payload []byte) string {
	engine := payload[0]
	paramId := payload[1]
	fmt.Println("RES:" + hex.Dump(payload))

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

/*
request:
00000000  06 0b 00 16 00 04                                 |......|


response:
00000000  06 0b 00 16 00 00 00 00  00 00 00 00              |............|


request:
00000000  06 0d 00 26 00 04                                 |...&..|

response:
00000000  06 0d 00 26 00 00 00 00  00 00 00 00              |...&........|



*/
