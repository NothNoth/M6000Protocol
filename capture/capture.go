package capture

import (
	"log"
	"m6kparse/tcpparser"
	"m6kparse/udpparser"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type Capture struct {
	logs      *log.Logger
	udpParser *udpparser.UDPParser
	tcpParser *tcpparser.TCPParser
}

func New(logs *log.Logger, iconIP string, frameIP string) *Capture {
	var cap Capture

	cap.logs = logs
	cap.udpParser = udpparser.New(iconIP, frameIP, cap.logs)
	cap.tcpParser = tcpparser.New(iconIP, frameIP, cap.logs)

	return &cap
}

func (cap *Capture) ReadLive(networkInterface string) error {

	h, err := pcap.OpenLive(networkInterface, 1500, true, 1*time.Millisecond)
	if err != nil {
		return err
	}

	return cap.capturePackets(h)
}

func (cap *Capture) ReadPcap(pcapFile string) error {

	h, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		return err
	}

	return cap.capturePackets(h)
}

func (cap *Capture) capturePackets(h *pcap.Handle) error {

	packetSource := gopacket.NewPacketSource(h, h.LinkType())
	for packet := range packetSource.Packets() {
		ipLayer := packet.Layer(layers.LayerTypeIPv4)

		if ipLayer != nil {
			var tcp *layers.TCP
			var udp *layers.UDP
			ip, _ := ipLayer.(*layers.IPv4)

			udpLayer := packet.Layer((layers.LayerTypeUDP))
			if udpLayer != nil {
				udp, _ = udpLayer.(*layers.UDP)
			}
			tcpLayer := packet.Layer((layers.LayerTypeTCP))
			if tcpLayer != nil {
				tcp, _ = tcpLayer.(*layers.TCP)
			}

			if tcp != nil {
				cap.tcpParser.Parse(packet, ip, tcp)
			} else if udp != nil {
				cap.udpParser.Parse(packet, ip, udp)
			}
		}
	}
	return nil
}
