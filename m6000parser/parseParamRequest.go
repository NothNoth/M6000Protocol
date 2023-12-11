package m6000parser

import "fmt"

type ParamRequest struct {
}

func (cmd *ParamRequest) Parse(payload []byte) string {
	engine := payload[0]
	paramId := payload[1]
	//unkA := payload[2]
	//unkB := payload[3]
	//count := midiTwoBytesTo14Bits(payload[4], payload[5])
	return fmt.Sprintf("Param request (Eng %d - Param %d)", engine, paramId)
}
