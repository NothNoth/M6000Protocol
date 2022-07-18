package tcpparser

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"m6kparse/common"
	"m6kparse/midi"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type TCPParser struct {
	iconIP               string
	frameIP              string
	midiParser           *midi.MIDI
	iconToFrameTruncated []byte
	frameToIconTruncated []byte
}

func New(iconIP string, frameIP string) *TCPParser {
	var p TCPParser

	p.iconIP = iconIP
	p.frameIP = frameIP
	p.midiParser = midi.New()
	return &p
}

func (p *TCPParser) Parse(packet gopacket.Packet, ip *layers.IPv4, tcp *layers.TCP) {

	if (ip.SrcIP.String() == p.frameIP) && (ip.DstIP.String() == p.iconIP) {
		p.parseBlocks(tcp.Payload, common.FrameToIcon)
	} else if (ip.SrcIP.String() == p.iconIP) && (ip.DstIP.String() == p.frameIP) {
		p.parseBlocks(tcp.Payload, common.IconToFrame)
	}
}

func (p *TCPParser) parseBlocks(payload []byte, d common.Direction) {
	offs := 0
	msg := 1

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
		if (d == common.IconToFrame) && (len(p.iconToFrameTruncated) != 0) {
			midiData = append(p.iconToFrameTruncated, midiData...)
		} else if (d == common.FrameToIcon) && (len(p.frameToIconTruncated) != 0) {
			midiData = append(p.frameToIconTruncated, midiData...)
		}
		fmt.Println(hex.Dump(midiData))
		p.midiParser.Parse(midiData, d)
		offs += size

		//Reached the end
		if offs == len(payload) {
			return
		}
		msg++
	}
}
