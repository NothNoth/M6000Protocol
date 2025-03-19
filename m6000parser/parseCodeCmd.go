package m6000parser

import (
	"encoding/hex"
	"fmt"
)

type CodeCmd struct {
}

func (cmd *CodeCmd) Parse(payload []byte) string {
	//header := payload[:2]

	fmt.Println("Full payload:")
	fmt.Println(hex.Dump(payload))

	decoded := make([]byte, len(payload)-3)
	idx := 0
	for i := 2; i < len(payload)-2; i += 2 {
		a := payload[i]
		b := payload[i+1]
		val := ((uint16(a)) << 4) | uint16(b)
		decoded[idx] = byte(val)
		idx++
	}

	return "Licence submit: " + string(decoded)
}
