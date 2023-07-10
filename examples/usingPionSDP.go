package main

import (
	"net"

	"github.com/openaudiocollective/sap"
	"github.com/pion/sdp"
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

	/*
		v=0
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
		a=mediaclk:direct=142410716
	*/

	inf := sdp.Information("2 channels: TxChan 0, TxChan 1")
	rng := 32
	sess := sdp.SessionDescription{
		Version: 0,
		Origin: sdp.Origin{
			Username:       "-",
			SessionID:      1423986,
			SessionVersion: 1423994,
			NetworkType:    "IN",
			AddressType:    "IP4",
			UnicastAddress: "169.254.98.63",
		},
		SessionName:        "AOIP44-serial-1614 : 2",
		SessionInformation: &inf,
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN",
			AddressType: "IP4",
			Address: &sdp.Address{
				IP:    net.ParseIP("239.65.45.154"),
				Range: &rng,
			},
		},
		TimeDescriptions: []sdp.TimeDescription{
			sdp.TimeDescription{
				Timing: sdp.Timing{StartTime: 0, StopTime: 0},
			},
		},
		Attributes: []sdp.Attribute{
			sdp.Attribute{
				Key:   "keywds",
				Value: "Dante",
			},
		},
		MediaDescriptions: []*sdp.MediaDescription{
			&sdp.MediaDescription{
				MediaName: sdp.MediaName{Media: "audio", Port: sdp.RangedPort{Value: 5004}, Protos: []string{"RTP/AVP"}, Formats: []string{"97"}},
				Attributes: []sdp.Attribute{
					sdp.Attribute{
						Key: "recvonly",
					},
					sdp.Attribute{
						Key:   "rtpmap",
						Value: "96 L24/48000/2",
					},
					sdp.Attribute{
						Key:   "ptime",
						Value: "1",
					},
					sdp.Attribute{
						Key:   "ts-refclk",
						Value: "ptp=IEEE1588-2008:00-00-00-FF-FE-00-00-00:0",
					},
					sdp.Attribute{
						Key:   "mediaclk",
						Value: "direct=142410716`",
					},
				},
			},
		},
	}

	payload := []byte(sess.Marshal())

	pckt, err := sap.NewPacket(payload, *raddr)
	if err != nil {
		panic(err)
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
