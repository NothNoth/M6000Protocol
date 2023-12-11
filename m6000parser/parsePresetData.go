package m6000parser

import (
	"fmt"
)

type PresetData struct {
}

func (cmd *PresetData) Parse(payload []byte) string {
	presetNumber := uint16(payload[1])<<8 | uint16(payload[0])
	/*
		//Partially works
		//We have some actual preset strings but I can't determine a fixed preset size
			fmt.Println(hex.Dump(payload))
			idx := 0
			for {
				var presetNum int
				var presetName string
				fmt.Printf(">>> Preset starts at idx x%x\n", idx)
				fmt.Printf("Header %x %x %x\n", payload[idx], payload[idx+1], payload[idx+2])

				presetNum = int(midiTwoBytesTo14Bits(payload[idx], payload[idx+1]))
				idx += 2
				//Unknown
				idx++

				for i := idx; i < idx+40; i += 2 {
					nameVal := byte(uint16(payload[i])<<4 | uint16(payload[i+1]))
					presetName += string(nameVal)
				}
				idx += 40

				fmt.Println("Trailer:")
				fmt.Println(hex.Dump(payload[idx : idx+45]))
				//Crossfeed
				idx += 2
				//Reserved
				idx += 14
				idx += 29
				fmt.Printf("Preset %d : %s\n", presetNum, presetName)
				if idx >= len(payload) {
					break
				}
			}*/
	return fmt.Sprintf("Preset data %d", presetNumber)

}
