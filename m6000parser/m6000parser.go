package m6000parser

import (
	"encoding/binary"
	"log"
)

type BlockStatus int

const (
	StatusPacketInvalid    BlockStatus = iota
	StatusPacketSplit      BlockStatus = iota
	StatusPacketSplitFinal BlockStatus = iota
	StatusPacketFull       BlockStatus = iota
)

type Result struct {
	Status BlockStatus
}

type blockData struct {
	startPacketNumber int
	endPacketNumber   int
	data              []byte
}

type M6000Parser struct {
	logs *log.Logger

	partialStartPacketNumber int
	partialBlockSize         int
	partialBlockData         []byte

	blockList []blockData
}

func New(logs *log.Logger) *M6000Parser {
	var m6p M6000Parser
	m6p.logs = logs

	m6p.partialBlockSize = 0
	return &m6p
}

func (m6p *M6000Parser) PushPacket(packetNumber int, data []byte) Result {

	m6p.logs.Printf("> Push packet #%d with %d bytes of data\n", packetNumber, len(data))

	//Not block in progress, start with a fresh one
	if m6p.partialBlockSize == 0 {
		dataStartIdx := 0

		for {
			version := int(binary.BigEndian.Uint16(data[dataStartIdx : dataStartIdx+2]))
			blockSize := int(binary.BigEndian.Uint16(data[dataStartIdx+2 : dataStartIdx+4]))
			m6p.logs.Printf(">> Block v%d of size %d\n", version, blockSize)

			if version != 0x02 {
				m6p.logs.Println(">> Version seems invalid, skipping")
				return Result{Status: StatusPacketInvalid}
			}
			if dataStartIdx+int(blockSize) <= len(data) {
				//Can extract a complete block
				b := data[dataStartIdx+4 : dataStartIdx+4+blockSize]
				m6p.pushBlock(packetNumber, packetNumber, b)
				dataStartIdx = dataStartIdx + 4 + blockSize

				//Reached the end
				if dataStartIdx == len(data) {
					return Result{Status: StatusPacketFull}
				}
			} else {
				//Cannot extract, memorize partial data and stop here
				m6p.logs.Println(">> Block is truncated, saving for later")
				m6p.partialBlockData = data[dataStartIdx+4:]
				m6p.partialBlockSize = blockSize
				m6p.partialStartPacketNumber = packetNumber
				return Result{Status: StatusPacketSplit}
			}
		}
	}

	m6p.partialBlockData = append(m6p.partialBlockData, data...)

	//Still not enough
	if len(m6p.partialBlockData) < m6p.partialBlockSize {
		m6p.partialBlockData = append(m6p.partialBlockData, data...)
		m6p.logs.Printf(">> Block still truncated (%d required / %d available)\n", m6p.partialBlockSize, len(m6p.partialBlockData))
		return Result{Status: StatusPacketSplit}
	}

	//Got complete block, extract
	m6p.logs.Println(">> Block truncated fully available")
	b := m6p.partialBlockData[:m6p.partialBlockSize]
	m6p.pushBlock(m6p.partialStartPacketNumber, packetNumber, b)

	remainingData := m6p.partialBlockData[m6p.partialBlockSize:]
	m6p.partialBlockData = []byte{}
	m6p.partialBlockSize = 0
	m6p.partialStartPacketNumber = 0

	//Processed all data
	if len(remainingData) == 0 {
		return Result{Status: StatusPacketSplitFinal}
	}

	return m6p.PushPacket(packetNumber, remainingData)
}

func (m6p *M6000Parser) pushBlock(packetStartNumber int, packetEndNumber int, block []byte) {
	m6p.blockList = append(m6p.blockList, blockData{startPacketNumber: packetStartNumber, endPacketNumber: packetEndNumber, data: block})
}
