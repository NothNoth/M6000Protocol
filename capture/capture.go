package capture

import (
	"m6kparse/tcpparser"
	"m6kparse/udpparser"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func ReadLive(networkInterface string, iconIP string, frameIP string) error {

	h, err := pcap.OpenLive(networkInterface, 1500, true, 1*time.Millisecond)
	if err != nil {
		return err
	}

	return capturePackets(h, iconIP, frameIP)
}

func ReadPcap(pcapFile string, iconIP string, frameIP string) error {

	h, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		return err
	}

	return capturePackets(h, iconIP, frameIP)
}

func capturePackets(h *pcap.Handle, iconIP string, frameIP string) error {

	udpParser := udpparser.New(iconIP, frameIP)
	tcpParser := tcpparser.New(iconIP, frameIP)

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
				tcpParser.Parse(packet, ip, tcp)
			} else if udp != nil {
				udpParser.Parse(packet, ip, udp)
			}
		}
	}
	return nil
}
