package sap

import (
	"encoding/binary"
	"io"
	"mime"
	"net"
)

// Header represents an SAP packet header
type Header struct {
	// The version number field MUST be set to 1
	Version uint8

	// If the A bit is 0, the originating source field contains a 32-bit IPv4 address.
	// If the A bit is 1, the originating source contains a 128-bit IPv6 address.
	AddressType AddressType

	// SAP announcers MUST set this to 0, SAP listeners MUST ignore the contents of this field.
	Reserved uint8

	// If this bit is set to 0 this is a session announcement packet
	// If this bit is set to 1 this is a session deletion packet.
	MessageType MessageType

	// If the encryption bit is set to 1, the payload of the SAP packet is encrypted.
	// If this bit is 0 the packet is not encrypted.
	Encrypted uint8

	// If the compressed bit is set to 1, the payload is
	// compressed using the zlib compression algorithm (https://datatracker.ietf.org/doc/html/rfc2974#ref-3)
	Compressed uint8

	// An 8 bit unsigned quantity giving the number of 32 bit words following the main SAP header that contain aauthentication data.
	// If it is zero, no authentication header is present.
	AuthenticationLength uint8

	// Digital signature of the packet, with length as specified by the authentication length header field.
	AuthenticationData []uint32

	// A 16 bit quantity that, used in combination with the originating source, provides a globally unique identifier
	// indicating the precise version of this announcement.
	//
	// It MUST be unique for each session announced by a particular SAP announcer
	// and it MUST be changed if the session description is modified
	// (and a session deletion message SHOULD be sent for the old version of the session).
	//
	// SAP listeners MAY silently discard messages if the message
	// identifier hash is set to zero.
	MessageIDHash uint16

	// This gives the IP address of the original source
	// of the message.  This is an IPv4 address if the A field is set to
	// zero, else it is an IPv6 address. The address is stored in
	// network byte order.
	OriginatingSource net.IP

	// The payload type field is a MIME content type specifier, describing the format of the payload.
	//
	// This is a variable length ASCII text string, followed by a single zero byte (ASCII NUL).
	// The payload type SHOULD be included in all packets.
	// If the payload type is `application/sdp' both the payload type and its terminating zero byte MAY be omitted
	//
	// Technically, it is part of the Payload, but makes more sense to parse it with the rest of the header
	PayloadType string
}

type MessageType uint8

const (
	Announcement MessageType = 0
	Deletion     MessageType = 1
)

type AddressType uint8

const (
	IPv4 AddressType = 0
	IPv6 AddressType = 0
)

const (
	versionShift     = 5    // Number of bits to shift for the version bit
	addressShift     = 4    // Number of bits to shift for the address type bit
	reservedShift    = 3    // Number of bits to shift for the reserved bit
	messageTypeShift = 2    // Number of bits to shift for the message type bit
	encryptedShift   = 1    // Number of bits to shift for the encrypted bit
	compressedShift  = 0    // Number of bits to shift for the compressed bit
	oneBitMask       = 0x01 // Mask to shift 1 bit
)

// MarshalSize returns the size of the header once marshaled.
func (h Header) MarshalSize() int {
	// NOTE: Be careful to match the MarshalTo() method.
	// Flags (1 byte)
	size := 1

	// Authentication Length (1 byte)
	size += 1

	// Message Identifier Hash (2 bytes)
	size += 2

	// Authentication Data (variable)
	// Should be the same as int(h.AuthenticationLength) * 4
	size += len(h.AuthenticationData) * 4

	// Originating Source
	if h.AddressType == 0 {
		size += 4
	} else {
		size += 16
	}

	if h.PayloadType == "" {
		// payloadType is omitted
		return size
	}

	size += len([]byte(h.PayloadType))

	// Trailing zero after the Payload Type
	size++

	return size
}

// Unmarshal parses the passed byte slice and stores the result in the Header.
func (h *Header) Unmarshal(buf []byte) error {
	/*
	    0                   1                   2                   3
	    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   | V=1 |A|R|T|E|C|   auth len    |         msg id hash           |
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |                                                               |
	   :                originating source (32 or 128 bits)            :
	   :                                                               :
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |                    optional authentication data               |
	   :                              ....                             :
	   *-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*
	   |                      optional payload type                    |
	   +                                         +-+- - - - - - - - - -+
	   |                                         |0|     payload       |
	   + - - - - - - - - - - - - - - - - - - - - +-+- - - - - - - - - -|
	*/

	currentPosition := 0

	if len(buf[currentPosition:]) < 1 {
		return errBufTooSmallForFlags
	}

	// The first two bits are always zero
	// The third bit contains the Version Number
	h.Version = (buf[currentPosition] >> versionShift) & oneBitMask

	// The fourth bit is the address type bit
	h.AddressType = AddressType(buf[currentPosition]>>addressShift) & oneBitMask

	// The fifth bit is the reserved bit
	h.Reserved = (buf[currentPosition] >> reservedShift) & oneBitMask

	// The sixth bit is the message type bit
	h.MessageType = MessageType(buf[currentPosition]>>messageTypeShift) & oneBitMask

	// The seventh bit is the encrypted bit
	h.Encrypted = (buf[currentPosition] >> encryptedShift) & oneBitMask

	// The last bit is the compressed bit
	h.Compressed = (buf[currentPosition] >> compressedShift) & oneBitMask

	// First Byte
	currentPosition++

	// Authentication Length
	if len(buf[currentPosition:]) < 1 {
		return errBufTooSmallForAuthLength
	}
	h.AuthenticationLength = buf[currentPosition]
	currentPosition++

	// Message Id Hash
	if len(buf[currentPosition:]) < 2 {
		return errBufTooSmallForMsgIdHash
	}
	h.MessageIDHash = binary.BigEndian.Uint16(buf[currentPosition : currentPosition+2])
	currentPosition += 2

	// Originating Source
	switch h.AddressType {
	case 0: // Expecting IPv4
		if len(buf[currentPosition:]) < 4 {
			return errBufTooSmallForIPv4
		}

		h.OriginatingSource = net.IPv4(buf[currentPosition], buf[currentPosition+1], buf[currentPosition+2], buf[currentPosition+3])
		currentPosition += 4

	case 1: // Expecting IPv6
		if len(buf[currentPosition:]) < 16 {
			return errBufTooSmallForIPv6
		}

		h.OriginatingSource = net.IP(buf[currentPosition : currentPosition+16])
		currentPosition += 16
	}

	// Authentication Data
	if h.AuthenticationLength != 0 {
		if len(buf[currentPosition:]) < int(h.AuthenticationLength)*4 {
			return errBufTooSmallForAuthData
		}

		h.AuthenticationData = make([]uint32, h.AuthenticationLength)
		for i := 0; i < int(h.AuthenticationLength); i++ {
			h.AuthenticationData[i] = binary.BigEndian.Uint32(buf[currentPosition : currentPosition+4])
			currentPosition += 4 // each uint32 take up 4 bytes
		}
	}

	if len(buf[currentPosition:]) < 3 || (len(buf[currentPosition:]) >= 3 && string(buf[currentPosition:currentPosition+3]) == "v=0") {
		// whether there's is no payload or the payload type has been omitted
		// and we are already in the payload
		h.PayloadType = ""
	} else {
		// either there is a payload type in the header
		// or the payload type is "application/sdp" (implicit because it's omitted)
		// and the payload itself is not SDP (because it doesn't start with "v=0")

		i := 0
		for ; i < len(buf[currentPosition:]); i++ {
			if buf[currentPosition+i] == 0 { // looking for the trailing zero byte
				break
			}
		}

		if i == len(buf[currentPosition:]) && buf[currentPosition+(i-1)] != 0 {
			// we traversed the whole buffer but didnt find a trailing byte
			return errNoTrailingByteFound
		}

		mediaType, _, err := mime.ParseMediaType(string(buf[currentPosition : currentPosition+i])) // doesn't include the trailing zero
		if err != nil {
			// the string until the trailing zero is not a valid mime media type
			// this indicates the payload type has been omitted (thus being application/sdp) and we are already in the payload
			// since we already checked and the start of the payload is not "v=0", the payload is not of type "application/sdp"
			return err
		}

		// Payload Type
		h.PayloadType = mediaType
	}

	return nil
}

// Marshal serializes the header into bytes.
func (h Header) Marshal() (buf []byte, err error) {
	buf = make([]byte, h.MarshalSize())

	n, err := h.MarshalTo(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// MarshalTo serializes the header and writes to the buffer.
// It returns the number of bytes read n and any error.
func (h Header) MarshalTo(buf []byte) (n int, err error) {
	/*
	    0                   1                   2                   3
	    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   | V=1 |A|R|T|E|C|   auth len    |         msg id hash           |
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |                                                               |
	   :                originating source (32 or 128 bits)            :
	   :                                                               :
	   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	   |                    optional authentication data               |
	   :                              ....                             :
	   *-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*
	   |                      optional payload type                    |
	   +                                         +-+- - - - - - - - - -+
	   |                                         |0|     payload       |
	   + - - - - - - - - - - - - - - - - - - - - +-+- - - - - - - - - -|
	*/

	size := h.MarshalSize()
	if size > len(buf) {
		return 0, io.ErrShortBuffer
	}

	// This is the number of bytes marshalled
	currentPosition := 0

	// The first two bits are always zero
	// The third bit contains the Version Number
	buf[currentPosition] = (h.Version << versionShift)

	// The fourth bit is the address type bit
	buf[currentPosition] |= byte((h.AddressType << addressShift))

	// The fifth bit is the reserved bit
	buf[currentPosition] |= (h.Reserved << reservedShift)

	// The sixth bit is the message type bit
	buf[currentPosition] |= byte((h.MessageType << messageTypeShift))

	// The seventh bit is the encrypted bit
	buf[currentPosition] |= (h.Encrypted << encryptedShift)

	// The last bit is the compressed bit
	buf[currentPosition] |= (h.Compressed << compressedShift)

	// First byte
	currentPosition++

	// Authentication Length
	buf[currentPosition] = h.AuthenticationLength
	currentPosition++

	// Message Id Hash
	binary.BigEndian.PutUint16(buf[currentPosition:], h.MessageIDHash)
	currentPosition += 2

	// Originating Source
	if h.OriginatingSource.To4() != nil {
		copy(buf[currentPosition:], h.OriginatingSource.To4())
		currentPosition += 4
	} else if h.OriginatingSource.To16() != nil {
		copy(buf[currentPosition:], h.OriginatingSource.To16())
		currentPosition += 16
	} else {
		return 0, errInvalidIPOnHeader
	}

	// Authentication Data
	for _, data := range h.AuthenticationData {
		binary.BigEndian.PutUint32(buf[currentPosition:currentPosition+4], data)
		currentPosition += 4 // each uint32 take up 4 bytes
	}

	// Payload Type
	if h.PayloadType != "" {
		payloadTypeBytes := []byte(h.PayloadType)
		for i := range payloadTypeBytes {
			buf[currentPosition+i] = payloadTypeBytes[i]
		}
		currentPosition += len(payloadTypeBytes)

		// Trailing zero after the Payload Type
		buf[currentPosition] = 0
		currentPosition++
	}

	// PayloadType has been omitted
	// which indicates it is "application/sdp"
	return currentPosition, nil
}

// Clone returns a deep copy of h.
func (h Header) Clone() Header {
	clone := h

	// Deep copy of AuthenticationData slice
	if h.AuthenticationData != nil {
		clone.AuthenticationData = make([]uint32, len(h.AuthenticationData))
		copy(clone.AuthenticationData, h.AuthenticationData)
	}

	// Deep copy of OriginatingSource IP (net.IP is a slice internally)
	if h.OriginatingSource != nil {
		clone.OriginatingSource = make(net.IP, len(h.OriginatingSource))
		copy(clone.OriginatingSource, h.OriginatingSource)
	}

	// Copy string
	clone.PayloadType = h.PayloadType

	return clone
}
