package m6000parser

import (
	"encoding/binary"
	"log"
	"m6kparse/common"
	"m6kparse/midi"
)

type BlockStatus int

const (
	StatusPacketInvalid    BlockStatus = iota
	StatusPacketSplit      BlockStatus = iota
	StatusPacketSplitFinal BlockStatus = iota
	StatusPacketFull       BlockStatus = iota
)

type Result struct {
	Status      BlockStatus
	Description []string
}

type blockData struct {
	startPacketNumber int
	endPacketNumber   int
	data              []byte
}

type M6000Parser struct {
	logs *log.Logger
	dir  common.Direction
	midi *midi.MIDI

	//Split block management
	partialStartPacketNumber int
	partialBlockSize         int
	partialBlockData         []byte

	//Split Sysex management
	partialSysex []byte

	//
	blockList  []blockData
	cmdParsers map[byte]CmdParser
}

type CmdParser interface {
	Parse(payload []byte) string
}

func New(logs *log.Logger, dir common.Direction) *M6000Parser {
	var m6p M6000Parser
	m6p.logs = logs
	m6p.dir = dir
	m6p.partialStartPacketNumber = 0
	m6p.partialBlockSize = 0

	m6p.cmdParsers = make(map[byte]CmdParser)
	m6p.cmdParsers[SYXTYPE_CODECMD] = new(CodeCmd)
	m6p.cmdParsers[SYXTYPE_CODECMD_RESPONSE] = new(CodeCmdResponse)
	m6p.cmdParsers[SYXTYPE_PARAMREQUEST] = new(ParamRequest)
	m6p.cmdParsers[SYXTYPE_PARAMDATA] = new(ParamResponse)
	m6p.cmdParsers[SYXTYPE_PRESETREQUEST] = new(PresetRequest)
	m6p.cmdParsers[SYXTYPE_PRESETDATA] = new(PresetData)
	return &m6p
}

func (m6p *M6000Parser) PushPacket(packetNumber int, data []byte) Result {
	var result Result
	//m6p.logs.Println(hex.Dump(data))

	//m6p.logs.Printf("> Push packet #%d with %d bytes of data\n", packetNumber, len(data))

	//Not block in progress, start with a fresh one
	if m6p.partialBlockSize == 0 {
		dataStartIdx := 0

		for {
			version := int(binary.BigEndian.Uint16(data[dataStartIdx : dataStartIdx+2]))
			blockSize := int(binary.BigEndian.Uint16(data[dataStartIdx+2 : dataStartIdx+4]))
			//m6p.logs.Printf(">> Block v%d of size %d\n", version, blockSize)

			if version != 0x02 {
				//m6p.logs.Println(">> Version seems invalid, skipping")
				result.Status = StatusPacketInvalid
				return result
			}
			if dataStartIdx+int(blockSize) <= len(data) {
				//Can extract a complete block
				b := data[dataStartIdx+4 : dataStartIdx+4+blockSize]
				result.Description = append(result.Description, m6p.pushBlock(packetNumber, packetNumber, b))
				dataStartIdx = dataStartIdx + 4 + blockSize
				//m6p.logs.Println(">> Full block read")

				//Reached the end
				if dataStartIdx == len(data) {
					result.Status = StatusPacketFull
					return result
				}
			} else {
				//Cannot extract, memorize partial data and stop here
				//m6p.logs.Println(">> Block is truncated, saving for later")
				m6p.partialBlockData = data[dataStartIdx+4:]
				m6p.partialBlockSize = blockSize
				m6p.partialStartPacketNumber = packetNumber
				result.Status = StatusPacketSplit
				return result
			}
		}
	}

	m6p.partialBlockData = append(m6p.partialBlockData, data...)

	//Still not enough
	if len(m6p.partialBlockData) < m6p.partialBlockSize {
		m6p.partialBlockData = append(m6p.partialBlockData, data...)
		//m6p.logs.Printf(">> Block still truncated (%d required / %d available)\n", m6p.partialBlockSize, len(m6p.partialBlockData))
		result.Status = StatusPacketSplit
		return result
	}

	//Got complete block, extract
	//m6p.logs.Println(">> Block truncated fully available")
	b := m6p.partialBlockData[:m6p.partialBlockSize]
	result.Description = append(result.Description, m6p.pushBlock(m6p.partialStartPacketNumber, packetNumber, b))

	remainingData := m6p.partialBlockData[m6p.partialBlockSize:]
	m6p.partialBlockData = []byte{}
	m6p.partialBlockSize = 0
	m6p.partialStartPacketNumber = 0

	//Processed all data
	if len(remainingData) == 0 {
		result.Status = StatusPacketSplitFinal
		return result
	}

	return m6p.PushPacket(packetNumber, remainingData)
}

func (m6p *M6000Parser) pushBlock(packetStartNumber int, packetEndNumber int, block []byte) string {
	var b blockData

	b.startPacketNumber = packetStartNumber
	b.endPacketNumber = packetEndNumber
	b.data = block
	m6p.blockList = append(m6p.blockList, b)

	return m6p.parseBlock(b)
}
