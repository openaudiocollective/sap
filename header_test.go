package sap

import (
	"net"
	"reflect"
	"testing"
)

// TestHeaderMarshalSize checks the size of the marshaled header.
func TestHeaderMarshalSize(t *testing.T) {
	testCases := []struct {
		name       string
		mockHeader Header
		want       int
	}{
		{
			name:       "Test 1: Everything Default",
			mockHeader: CreateMockHeader(Header{}),
			want:       8,
		},
		{
			name: "Test 2: IPv6",
			mockHeader: CreateMockHeader(Header{
				AddressType: 1,
			}),
			want: 20,
		},
		{
			name: "Test 3: AuthenticationLength = 1",
			mockHeader: CreateMockHeader(Header{
				AuthenticationLength: 1,
			}),
			want: 12,
		},
		{
			name: "Test 4: AuthenticationLength = 5",
			mockHeader: CreateMockHeader(Header{
				AuthenticationLength: 5,
			}),
			want: 28,
		},
		{
			name: "Test 5: IPv6 and AuthenticationLength = 5",
			mockHeader: CreateMockHeader(Header{
				AddressType:          1,
				AuthenticationLength: 5,
			}),
			want: 40,
		},
		{
			name: "Test 6: Payload Type application/sdp",
			mockHeader: CreateMockHeader(Header{
				PayloadType: "application/sdp",
			}),
			want: 24,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.mockHeader.MarshalSize()
			if tc.want != got {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestHeaderUnmarshalErrors(t *testing.T) {
	testCases := []struct {
		name          string
		input         []byte
		expectedError error
	}{
		{
			name:          "BufTooSmallForFlags",
			input:         make([]byte, 0),
			expectedError: errBufTooSmallForFlags,
		},
		{
			name:          "BufTooSmallForAuthLength",
			input:         make([]byte, 1),
			expectedError: errBufTooSmallForAuthLength,
		},
		{
			name:          "BufTooSmallForMsgIdHash",
			input:         make([]byte, 2),
			expectedError: errBufTooSmallForMsgIdHash,
		},
		{
			name:          "BufTooSmallForForIPv4",
			input:         make([]byte, 4),
			expectedError: errBufTooSmallForIPv4,
		},
		{
			name: "BufTooSmallForForIPv6",
			// The address type bit is set to 1 and this byte slice has 16 bytes which is not enough for IPv6 address
			input:         []byte{0x30, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectedError: errBufTooSmallForIPv6,
		},
		{
			name:          "BufTooSmallForAuthDataWithIPv4",
			input:         []byte{0x20, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectedError: errBufTooSmallForAuthData,
		},
		{
			name: "BufTooSmallForAuthDataWithIPv6",
			// The address type bit is set to 1 and this byte slice has 20 bytes which is enough for IPv6 address but not to AuthData
			input:         []byte{0x30, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expectedError: errBufTooSmallForAuthData,
		},
		{
			name: "NoTrailingByteFound",
			// There's more than one byte after the AuthenticationData and but none is 0 (the trailing byte)
			input:         []byte{0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x01},
			expectedError: errNoTrailingByteFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHeader := CreateMockHeader(Header{})
			err := mockHeader.Unmarshal(tc.input)
			if err != tc.expectedError {
				t.Errorf("Expected error %v, but got %v", tc.expectedError, err)
			}
		})
	}
}

// TestHeaderMarshalAndUnmarshal checks that a header can be marshaled and then unmarshaled to its original state.
func TestHeaderMarshalAndUnmarshal(t *testing.T) {
	testCases := []struct {
		name       string
		mockHeader Header
	}{
		{
			name:       "Test 1: Everything Default",
			mockHeader: CreateMockHeader(Header{}),
		},
		{
			name: "Test 2: IPv6",
			mockHeader: CreateMockHeader(Header{
				AddressType: 1,
			}),
		},
		{
			name: "Test 3: AuthenticationLength = 1",
			mockHeader: CreateMockHeader(Header{
				AuthenticationLength: 1,
			}),
		},
		{
			name: "Test 4: AuthenticationLength = 5",
			mockHeader: CreateMockHeader(Header{
				AuthenticationLength: 5,
			}),
		},
		{
			name: "Test 5: IPv6 and AuthenticationLength = 5",
			mockHeader: CreateMockHeader(Header{
				AddressType:          1,
				AuthenticationLength: 5,
			}),
		},
		{
			name: "Test 6: Payload Type application/sdp",
			mockHeader: CreateMockHeader(Header{
				PayloadType: "application/sdp",
			}),
		},
		{
			name: "Test 7: Payload Type application/json",
			mockHeader: CreateMockHeader(Header{
				PayloadType: "application/json",
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.mockHeader.Marshal()
			if err != nil {
				t.Errorf("Marshal failed with error: %v", err)
			}

			h2 := CreateMockHeader(tc.mockHeader)

			err = h2.Unmarshal(data)
			if err != nil {
				t.Errorf("Unmarshal failed with error: %v", err)
			}

			if !reflect.DeepEqual(tc.mockHeader, h2) {
				t.Error("original and unmarshalled headers do not match")
			}
		})
	}
}

// TestHeaderMarshalTo checks that the header can be serialized and written to a byte slice.
func TestHeaderMarshalTo(t *testing.T) {
	testCases := []struct {
		name       string
		mockHeader Header
	}{
		{
			name:       "Test 1: Everything Default",
			mockHeader: CreateMockHeader(Header{}),
		},
		{
			name: "Test 2: IPv6",
			mockHeader: CreateMockHeader(Header{
				AddressType: 1,
			}),
		},
		{
			name: "Test 3: AuthenticationLength = 1",
			mockHeader: CreateMockHeader(Header{
				AuthenticationLength: 1,
			}),
		},
		{
			name: "Test 4: AuthenticationLength = 5",
			mockHeader: CreateMockHeader(Header{
				AuthenticationLength: 5,
			}),
		},
		{
			name: "Test 5: IPv6 and AuthenticationLength = 5",
			mockHeader: CreateMockHeader(Header{
				AddressType:          1,
				AuthenticationLength: 5,
			}),
		},
		{
			name: "Test 6: Payload Type application/sdp",
			mockHeader: CreateMockHeader(Header{
				PayloadType: "application/sdp",
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bufSize := tc.mockHeader.MarshalSize()
			buf := make([]byte, bufSize)

			n, err := tc.mockHeader.MarshalTo(buf)
			if err != nil {
				t.Errorf("MarshalTo failed with error: %v", err)
			}

			if n != len(buf) {
				t.Errorf("expected to write %d bytes, wrote %d", len(buf), n)
			}
		})
	}
}

func TestHeaderMarshalTo_InvalidIP(t *testing.T) {
	h := CreateMockHeader(Header{
		OriginatingSource: net.IP{0, 0, 0}, // invalid IP address, not 4 or 16 bytes
	})

	_, err := h.MarshalTo(make([]byte, 1024))
	if err == nil {
		t.Fatalf("Expected error, but got nil")
	}
}

// TestHeaderClone checks that cloning a header returns a deep copy.
func TestHeaderClone(t *testing.T) {
	h1 := CreateMockHeader(Header{
		AuthenticationLength: 5,
		AuthenticationData:   make([]uint32, 5),
	})

	h2 := h1.Clone()
	if !reflect.DeepEqual(&h1, &h2) {
		t.Error("Clone did not create a deep copy")
	}
}

func CreateMockHeader(h Header) Header {
	// Any fields with their zero values (0 for integers, nil for slices, and so on) will be filled in with default values.
	h.Version = 1

	if h.AddressType == 0 {
		if h.OriginatingSource == nil {
			h.OriginatingSource = net.ParseIP("192.0.2.1")
		}
	} else {
		if h.OriginatingSource == nil {
			h.OriginatingSource = net.ParseIP("2001:db8::68")
		}
	}

	if h.Reserved == 0 {
		h.Reserved = 0
	}

	if h.MessageType == 0 {
		h.MessageType = 0
	}

	if h.Encrypted == 0 {
		h.Encrypted = 0
	}

	if h.Compressed == 0 {
		h.Compressed = 0
	}

	if h.AuthenticationLength == 0 {
		h.AuthenticationLength = 0
	}

	h.AuthenticationData = make([]uint32, h.AuthenticationLength)

	// Fill the AuthenticationData with some mock data
	for i := range h.AuthenticationData {
		h.AuthenticationData[i] = uint32(i + 1)
	}

	if h.MessageIDHash == 0 {
		h.MessageIDHash = 12345
	}

	if h.PayloadType == "" {
		h.PayloadType = ""
	}

	return h
}
