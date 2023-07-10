package sap

import (
	"fmt"
)

// Packet represents an SAP Packet
type Packet struct {
	Header
	Payload []byte
}

// String helps with debugging by printing packet information in a readable way
func (p Packet) String() string {
	out := "SAP PACKET:\n"

	out += fmt.Sprintf("\tVersion: %v\n", p.Version)
	out += fmt.Sprintf("\tAddressType: %v\n", p.AddressType)
	out += fmt.Sprintf("\tReserved: %d\n", p.Reserved)
	out += fmt.Sprintf("\tMessageType: %d\n", p.MessageType)
	out += fmt.Sprintf("\tEncryptionBit: %d\n", p.Encrypted)
	out += fmt.Sprintf("\tCompressedBit: %d\n", p.Compressed)
	out += fmt.Sprintf("\tAuthenticationLength: %d\n", p.AuthenticationLength)
	out += fmt.Sprintf("\tAuthenticationData: %d\n", p.AuthenticationData)
	out += fmt.Sprintf("\tMessageIDHash: %d\n", p.MessageIDHash)
	out += fmt.Sprintf("\tOriginatingSource: %d\n", p.OriginatingSource)
	out += fmt.Sprintf("\tPayload Type: %s\n", p.PayloadType)
	out += fmt.Sprintf("\tPayload Length: %d\n", len(p.Payload))

	return out
}

// Marshal serializes the packet into bytes.
func (p Packet) Marshal() (buf []byte, err error) {
	buf = make([]byte, p.MarshalSize())

	n, err := p.MarshalTo(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

// MarshalTo serializes the packet and writes to the buffer.
// It returns the number of bytes read and any error.
func (p Packet) MarshalTo(buf []byte) (n int, err error) {
	headerSize := p.Header.MarshalSize()
	if len(buf) < headerSize {
		return 0, errBufTooSmallForHeader
	}

	// Add the hash to the Header if it doesn't have one
	if p.Header.MessageIDHash == 0 {
		p.Header.MessageIDHash = ComputeMsgIdHash(&p.Payload)
	}

	n, err = p.Header.MarshalTo(buf)
	if err != nil {
		return 0, err
	}

	// Make sure the buffer is large enough to hold the packet.
	if len(buf) < headerSize+len(p.Payload) {
		return 0, errBufTooSmallForPayload
	}

	payloadSize := copy(buf[n:], p.Payload)

	return headerSize + payloadSize, nil
}

// MarshalSize returns the size of the packet once marshaled.
func (p Packet) MarshalSize() int {
	return p.Header.MarshalSize() + len(p.Payload)
}

// Unmarshal parses the passed byte slice and stores the result in the Packet.
func (p *Packet) Unmarshal(buf []byte) error {
	err := p.Header.Unmarshal(buf)
	if err != nil {
		return err
	}

	headerSize := p.Header.MarshalSize()
	end := len(buf)
	if len(buf) > headerSize { // only slice if needed
		p.Payload = buf[headerSize:end]
	} else {
		// no payload
		return nil
	}
	return nil
}

// Clone returns a deep copy of p.
func (p Packet) Clone() *Packet {
	clone := &Packet{}
	clone.Header = p.Header.Clone()
	if p.Payload != nil {
		clone.Payload = make([]byte, len(p.Payload))
		copy(clone.Payload, p.Payload)
	}
	return clone
}
