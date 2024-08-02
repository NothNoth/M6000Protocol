package m6000parser

import (
	"encoding/hex"
	"fmt"
)

type ParamRequest struct {
}

/*
This is a 6 bytes message

Example:

00000000  06 00 00 24 00 02                                 |...$..|
00000000  06 00 00 28 00 04                                 |...(..|
00000000  06 0b 00 16 00 04                                 |......|
00000000  06 0d 00 26 00 04                                 |...&..|
00000000  06 78 00 00 00 2a                                 |.x...*|
00000000  06 79 00 00 00 2a                                 |.y...*|
00000000  06 7f 00 00 00 26                                 |.....&|
*/
func (cmd *ParamRequest) Parse(payload []byte) string {
	engine := payload[0]
	paramId := payload[1]
	//unkA := payload[2]
	//unkB := payload[3]
	//count := midiTwoBytesTo14Bits(payload[4], payload[5])

	fmt.Println("REQ:" + hex.Dump(payload))

	return fmt.Sprintf("Param request (Eng %d - Param %d)", engine, paramId)
}
