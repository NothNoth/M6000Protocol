package m6000parser

import (
	"encoding/hex"
)

type CodeCmdResponse struct {
}

func (cmd *CodeCmdResponse) Parse(payload []byte) string {
	if len(payload) != 3 {
		return "Licence response: invalid size" + hex.Dump(payload)
	}

	switch payload[2] {
	case 0x00:
		return "Licence response: Valid" //when sending already validated licence
	case 0x02:
		return "Licence response: too short" //When sending truncated
	case 0x03:
		return "Licence response: invalid checksum/licence num(?)" //when sending random stuff
	case 0x05:
		return "Licence response: invalid checksum" //when sending with last digit changed
	}

	return "Licence response: Unknown"

}
