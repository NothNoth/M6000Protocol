# M6k Parse

A network packet parser for the TC Electronic M6000.

## Introduction

This project is a packet parser, analyzing TCP frames sent between the M6000 Mainframe and the Icon remote.
This is a Work In Progress (WIP).

## Network traffic

By default, the following IP addresses are used:

  - Icon: 192.168.1.125
  - Mainframe: the 192.168.1.249

The Icon unit is technically a Windows NT4 PC, thus it will use some well known Microsoft protocols by default such as NetBIOS.
Upon startup, the Icon declares itself on the NetBIOS UDP service.
Since we're not interested on this part of the traffic, this will be mostly ignored here.

In wireshark, we would typically the followinf filter in order to focus on the actual M6000 traffic:

    ip.addr == 192.168.1.125 and ((udp.port != 137 and udp.port != 138) or tcp)

## Discovery

### Probe

Upon boot, the Icon will send a broadcast UDP message in order to look for Mainframes on the network.
This UDP message is sent from 192.168.1.125 to 255.255.255.255 and uses the magic string 0x12345678 followed by the Icon device name: "TCIcon".

| Offset | Size | Example    | Description                                                     |
| ------ | ---- | ---------- | --------------------------------------------------------------- |
| 0      | 4    | 0x12345678 | Discovery magic                                                 |
| 4      | 10   | TCIcon    | Icon name, null terminated string                               |
| E      | 62   |            | Unknown data                                                    |

### Responses

The Mainframe will respond to this probe using 4 messages all starting with the same Magic (0x12345678):

  - 112255_M6000X_TEXT
  - 112255_M6000X_DISK
  - 112255_M6000X_MIDI
  - 112255_M6000X_METER

Where 112255 is the Mainframe serial number.
Note that the serial number is also found encoded as uint32 bigendian on the bytes 4 to 8 of those messages, just after the magic string.

| Offset | Size | Example             | Description                                                     |
| ------ | ---- | ------------------- | --------------------------------------------------------------- |
| 0      | 4    | 0x12345678          | Discovery magic                                                 |
| 4      | 4    | 0x0001B67F          | Device serial number                                            |
| 8      | 1    | 0x04                | (unsure) Total number of messages                               |
| 9      | 10   |                     | Unknown data                                                    |
| 13     | 1    | 0x01                | Message number (from 0x00 to Total number of messages)          |
| 14     | 20   | 112255_M6000X_MIDI  | Null terminated string (entry name)                             |
| 54     | 13   | TC Electronic S6000 | Null terminated string (device name)                            |

### Meter

After discovery, the Icon sends a Meter specific message:


| Offset | Size | Example    | Description                                                     |
| ------ | ---- | ---------- | --------------------------------------------------------------- |
| 0      | 4    | 0x12345678 | Discovery magic                                                 |
| 4      | 10   | METER PORT | Possible meter port config?                                     |
| E      | 62   |            | Unknown data                                                    |

## Timecodes

Once the Mainframe has been detected by the Icon, the only UDP traffic is the Timecodes that will start once the Mainframe is selected on the Icon.
The mainframe will send UDP timecodes to the Icon from port 1024 to port 1027.
Those timecodes formats have not been reversed.

## TCP Traffic

A TCP connection is established between the Mainframe and the Icon, on port 1026. This session is initiated by the Icon.
TCP payloads contains one of more "blocks" which seems to embed MIDI messages.
Each block has the following format:

| Offset | Size   | Example    | Description                                                     |
| ------ | ------ | ---------- | --------------------------------------------------------------- |
| 0      | 2      | 0x0002     | Version string                                                  |
| 2      | 2      | 0x0014     | Block size                                                      |
| 4      | <size> |            | MIDI Data                                                       |

One single payload may contain multiple blocks and blocks may be truncated and split over multiple payloads.

Misc notes:

  - Most blocks can be found at regular interval on the dumps

### First message

The very first TCP message is sent by the Mainframe to the Icon:

    00 02 00 03 ff 00 00                              .......

This can be decoded as follow:

  - 0002 : Version string
  - 0003 : Block length
  - FF0000 : MIDI Data

According to the MIDI specifications, 0xFF is a MIDI reset message.

### Others

Now that we know the M6000 uses MIDI SysEx messages we can try to extrapolate from other TC devices, such as the  M-One (https://mediadl.musictribe.com/download/software/tcelectronic/tc_electronic_m-one_xl_midi_sysex_specifications.pdf) or D-two (https://mediadl.musictribe.com/download/software/tcelectronic/tc_electronic_d-two_midi_sysex_specifications.pdf)

Other messages will typically look as follow:

    0000   00 02 00 0e f0 00 20 1f 00 46 47 06 7f 00 00 00   ...... ..FG.....
    0010   26 f7                                             &.

This can be decoded as follow:

  - 0002 : Version string
  - 000e : Block length
  - F0 ... F7 : MIDI SysEx data

The SysEx Message can be decoded as:

  - F0 : SysEx Start
  - 00 20 1F : "3 byte manufacturer ID for TC Electronic" (according to the D-Two spec)
  - 00 : System Exclusive device ID (User parameter set in "MIDI Setup Page")
  - 46 : M6000 model ID (D-Two was 0x45, M-One was 0x44)
  - 47 : Message Type
  - 06 7F 00 00 00 26 : Data
  - F7 : SysEnd Ends here

