package m6000parser

import (
	"fmt"
)

type PresetData struct {
}

func (cmd *PresetData) Parse(payload []byte) string {
	presetNumber := uint16(payload[1])<<8 | uint16(payload[0])
	/*
		idx := 3
		var str string
		var inString bool

		inString = true
		for {
			if idx+1 >= len(payload) {
				break
			}

			a := payload[idx]
			b := payload[idx+1]

			if inString {
				//End of string with double 0x00
				if a == 0 && b == 0 {
					inString = false
					fmt.Println("STR: ", str)
					str = ""

				} else {
					str += string(byte(uint16(a)<<4 | uint16(b)))
				}
				idx += 2
			} else {
				if (a != 0x00) && (a&0xF0 == 0x00) && (b != 0x00) && (b&0xF0 == 0x00) {
					inString = true
					str += string(byte(uint16(a)<<4 | uint16(b)))
					idx += 2
				} else {
					idx++
				}
			}
		}
	*/
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
