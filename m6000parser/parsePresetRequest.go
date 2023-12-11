package m6000parser

import "fmt"

type PresetRequest struct {
}

func (cmd *PresetRequest) Parse(payload []byte) string {
	presetNumber := uint16(payload[1])<<8 | uint16(payload[0])
	return fmt.Sprintf("Preset request %d", presetNumber)

}
