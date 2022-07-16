package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	tcFrameDetectionMagic = 0x12345678
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: " + os.Args[0] + " <pcap file> <frame IP> <icon IP>")
		fmt.Println("Example: " + os.Args[0] + " boot.pcap 192.168.1.126 192.168.1.125")
		return
	}

	pcapFileName := os.Args[1]
	frameIP := os.Args[2]
	iconIP := os.Args[3]

	h, err := pcap.OpenOffline(pcapFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

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

			if (ip.SrcIP.String() == frameIP) && (ip.DstIP.String() == iconIP) {
				if tcp != nil {
					parseFrameToIconTCP(packet, ip, tcp)
				} else if udp != nil {
					parseFrameToIconUDP(packet, ip, udp)
				} else {
					fmt.Println("Frame to Icon (unkown):")
					fmt.Println(packet)
				}
			} else if (ip.SrcIP.String() == iconIP) && (ip.DstIP.String() == frameIP) {
				if tcp != nil {
					parseIconToFrameTCP(packet, ip, tcp)
				} else if udp != nil {
					parseIconToFrameUDP(packet, ip, udp)
				} else {
					fmt.Println("Icon to Frame (unkown):")
					fmt.Println(packet)
				}
			} else if (ip.SrcIP.String() == iconIP) && (ip.DstIP[3] == 255) {
				parseIconToBroadcastUDP(packet, ip, udp)
			} else if ip.SrcIP.String() == frameIP {
				fmt.Println("Frame to unknown (" + ip.DstIP.String() + "):")
				fmt.Println(hex.Dump(ip.Payload))
			} else if ip.SrcIP.String() == iconIP {
				fmt.Println("Icon to unknown (" + ip.DstIP.String() + "):")
				fmt.Println(hex.Dump(ip.Payload))
			}
		}
	}
}
