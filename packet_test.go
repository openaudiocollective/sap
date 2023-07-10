package sap

import (
	"reflect"
	"testing"
)

// TestPacketMarshalSize checks the size of the marshaled packet.
func TestPacketMarshalSize(t *testing.T) {
	testCases := []struct {
		name       string
		mockPacket *Packet
		want       int
	}{
		{
			name:       "Test 1: Without Payload",
			mockPacket: CreateMockPacket(Packet{}),
			want:       8,
		},
		{
			name: "Test 2: With Payload",
			mockPacket: CreateMockPacket(Packet{
				Payload: []byte{0x10, 0x04},
			}),
			want: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.mockPacket.MarshalSize()
			if tc.want != got {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestPacketUnmarshalErrors(t *testing.T) {
	testCases := []struct {
		name          string
		input         []byte
		expectedError error
	}{
		{
			name:          "BufTooSmallForFlags",
			input:         make([]byte, 0), // Not enough bytes to unmarshal flags.
			expectedError: errBufTooSmallForFlags,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPacket := CreateMockPacket(Packet{})
			err := mockPacket.Unmarshal(tc.input)
			if err != tc.expectedError {
				t.Errorf("Expected error %v, but got %v", tc.expectedError, err)
			}
		})
	}
}

// TestPacketMarshalAndUnmarshal checks that a packet can be marshaled and then unmarshaled to its original state.
func TestPacketMarshalAndUnmarshal(t *testing.T) {
	testCases := []struct {
		name       string
		mockPacket *Packet
	}{
		{
			name:       "Test 1: Without Payload",
			mockPacket: CreateMockPacket(Packet{}),
		},
		{
			name: "Test 2: With Payload",
			mockPacket: CreateMockPacket(Packet{
				Payload: []byte{0x10, 0x04},
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.mockPacket.Marshal()
			if err != nil {
				t.Errorf("Marshal failed with error: %v", err)
			}

			p2 := CreateMockPacket(*tc.mockPacket)

			err = p2.Unmarshal(data)
			if err != nil {
				t.Errorf("Unmarshal failed with error: %v", err)
			}

			if !reflect.DeepEqual(tc.mockPacket, p2) {
				t.Error("original and unmarshalled packets do not match")
			}
		})
	}
}

// TestPacketMarshalTo checks that the packet can be serialized and written to a byte slice.
func TestPacketMarshalTo(t *testing.T) {
	testCases := []struct {
		name       string
		mockPacket *Packet
	}{
		{
			name:       "Test 1: Without Payload",
			mockPacket: CreateMockPacket(Packet{}),
		},
		{
			name: "Test 2: With Payload",
			mockPacket: CreateMockPacket(Packet{
				Payload: []byte{0x10, 0x04},
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bufSize := tc.mockPacket.MarshalSize()
			buf := make([]byte, bufSize)

			n, err := tc.mockPacket.MarshalTo(buf)
			if err != nil {
				t.Errorf("MarshalTo failed with error: %v", err)
			}

			if n != len(buf) {
				t.Errorf("expected to write %d bytes, wrote %d", len(buf), n)
			}
		})
	}
}

func TestHeaderMarshalToErrors(t *testing.T) {
	testCases := []struct {
		name          string
		input         []byte
		expectedError error
	}{
		{
			name:          "BufTooSmallForHeader",
			input:         make([]byte, 4),
			expectedError: errBufTooSmallForHeader,
		},
		{
			name:          "BufTooSmallForPayload",
			input:         make([]byte, 9),
			expectedError: errBufTooSmallForPayload,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockPacket := CreateMockPacket(Packet{
				Payload: []byte{0x10, 0x04},
			})
			_, err := mockPacket.MarshalTo(tc.input)
			if err != tc.expectedError {
				t.Errorf("Expected error %v, but got %v", tc.expectedError, err)
			}
		})
	}
}

// TestPacketClone checks that cloning a packet returns a deep copy.
func TestPacketClone(t *testing.T) {
	h1 := CreateMockPacket(Packet{})

	h2 := h1.Clone()
	if !reflect.DeepEqual(&h1, &h2) {
		t.Error("Clone did not create a deep copy")
	}
}

func CreateMockPacket(p Packet) *Packet {
	var newPayload []byte
	if p.Payload == nil {
		newPayload = nil
	} else {
		newPayload = make([]byte, len(p.Payload))
		copy(newPayload, p.Payload)
	}

	return &Packet{Header: CreateMockHeader(p.Header), Payload: newPayload}
}
