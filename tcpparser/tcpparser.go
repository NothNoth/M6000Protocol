package tcpparser

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"m6kparse/common"
	"m6kparse/midi"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type TCPParser struct {
	logs                 *log.Logger
	iconIP               string
	frameIP              string
	midiParser           *midi.MIDI
	iconToFrameTruncated []byte
	frameToIconTruncated []byte
}

func New(iconIP string, frameIP string, logs *log.Logger) *TCPParser {
	var p TCPParser

	p.iconIP = iconIP
	p.frameIP = frameIP
	p.logs = logs
	p.midiParser = midi.New(logs)
	return &p
}

func (p *TCPParser) Parse(packet gopacket.Packet, ip *layers.IPv4, tcp *layers.TCP) {

	if len(tcp.Payload) == 0 {
		return
	}
	p.logs.Println("************************************************************")
	p.logs.Printf("[TCP Packet] RAW Payload %d bytes (0x%d)\n", len(tcp.Payload), len(tcp.Payload))
	p.logs.Print("\n" + hex.Dump(tcp.Payload))

	if (ip.SrcIP.String() == p.frameIP) && (ip.DstIP.String() == p.iconIP) {
		p.logs.Println("-> Frame to icon (tcp)")
		p.parseBlocks(tcp.Payload, common.FrameToIcon)
	} else if (ip.SrcIP.String() == p.iconIP) && (ip.DstIP.String() == p.frameIP) {
		p.logs.Println("-> Icon to frame (tcp)")
		p.parseBlocks(tcp.Payload, common.IconToFrame)
	}
}

func (p *TCPParser) parseBlocks(payload []byte, d common.Direction) {
	offs := 0
	msg := 1

	//If data was truncated on previous packet, prepend saved
	if (d == common.IconToFrame) && (len(p.iconToFrameTruncated) != 0) {
		merged := append(p.iconToFrameTruncated, payload...)
		payload = make([]byte, len(merged))
		copy(payload, merged)
	} else if (d == common.FrameToIcon) && (len(p.frameToIconTruncated) != 0) {
		merged := append(p.frameToIconTruncated, payload...)
		payload = make([]byte, len(merged))
		copy(payload, merged)
	}

	for {
		//Not enough room for a complete block?
		if offs+4 > len(payload) {
			if d == common.IconToFrame {
				p.iconToFrameTruncated = make([]byte, len(payload)-offs)
				copy(p.iconToFrameTruncated, payload[offs:])
				return
			} else if d == common.FrameToIcon {
				p.frameToIconTruncated = make([]byte, len(payload)-offs)
				copy(p.frameToIconTruncated, payload[offs:])
				return
			}
			return
		}

		//version := int(binary.BigEndian.Uint16(payload[offs : offs+2]))
		offs += 2
		size := int(binary.BigEndian.Uint16(payload[offs : offs+2]))
		offs += 2

		if offs+size > len(payload) {
			//Not enough room
			offs -= 4
			if d == common.IconToFrame {
				p.iconToFrameTruncated = make([]byte, len(payload)-offs)
				copy(p.iconToFrameTruncated, payload[offs:])
				return
			} else if d == common.FrameToIcon {
				p.frameToIconTruncated = make([]byte, len(payload)-offs)
				copy(p.frameToIconTruncated, payload[offs:])
				return
			}
			return
		}

		midiData := payload[offs : offs+size]
		//p.logs.Println(hex.Dump(midiData))
		p.midiParser.Parse(midiData, d)
		offs += size

		//Reached the end
		if offs == len(payload) {
			return
		}
		msg++
	}
}
