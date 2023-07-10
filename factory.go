package sap

import "net"

type Option func(*Header)

// Set the MessageType of the Header struct
func WithMessageType(msgType MessageType) Option {
	return func(h *Header) {
		h.MessageType = msgType
	}
}

// Set the PayloadType of the Header struct
func WithPayloadType(payloadType string) Option {
	return func(h *Header) {
		h.PayloadType = payloadType
	}
}

// Factory function for creating a new SAP packet
// Creates an IPv4, unencrypted, uncompressed, unauthenticated, SAP/SDP announcement packet by default
func NewPacket(payload []byte, originatingSource net.UDPAddr, opts ...Option) (Packet, error) {
	header := Header{
		Version:              1,
		AddressType:          0,
		Reserved:             0,
		MessageType:          0,
		Encrypted:            0,
		Compressed:           0,
		AuthenticationLength: 0,
		AuthenticationData:   []uint32{},
		MessageIDHash:        ComputeMsgIdHash(payload),
		OriginatingSource:    originatingSource.IP,
		PayloadType:          "",
	}

	// Originating Source
	if originatingSource.IP.To4() != nil {
		header.AddressType = IPv4
	} else if originatingSource.IP.To16() != nil {
		header.AddressType = IPv6
	}

	// Apply the options to the header
	for _, opt := range opts {
		opt(&header)
	}

	return Packet{
		Header:  header,
		Payload: payload,
	}, nil
}
