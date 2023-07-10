package main

import (
	"net"

	"github.com/openaudiocollective/sap"
)

func main() {
	// Resolve the address
	raddr, err := net.ResolveUDPAddr("udp", "224.0.0.255:9875")
	if err != nil {
		panic(err)
	}

	// Establish the UDP connection
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		panic(err)
	}

	s := `v=0
o=- 1423986 1423994 IN IP4 169.254.98.63
s=AOIP44-serial-1614 : 2
c=IN IP4 239.65.45.154/32
t=0 0
a=keywds:Dante
m=audio 5004 RTP/AVP 97
i=2 channels: TxChan 0, TxChan 1
a=recvonly
a=rtpmap:96 L24/48000/2
a=ptime:1
a=ts-refclk:ptp=IEEE1588-2008:00-00-00-FF-FE-00-00-00:0
a=mediaclk:direct=142410716`

	payload := []byte(s)

	pckt := sap.Packet{
		Header: sap.Header{
			Version:              1,
			AddressType:          0,
			Reserved:             0,
			MessageType:          1,
			Encrypted:            0,
			Compressed:           0,
			AuthenticationLength: 0,
			AuthenticationData:   []uint32{},
			MessageIDHash:        sap.ComputeMsgIdHash(payload),
			OriginatingSource:    raddr.IP,
			PayloadType:          "",
		},
		Payload: payload,
	}

	mrshl, err := pckt.Marshal()
	if err != nil {
		panic(err)
	}

	_, err = conn.Write(mrshl)
	if err != nil {
		panic(err)
	}
}
