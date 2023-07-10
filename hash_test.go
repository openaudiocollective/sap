package sap

import "testing"

func TestValidateMsgIdHash(t *testing.T) {
	testCases := []struct {
		name    string
		payload []byte
		want    uint16
	}{
		{
			name:    "Test 1: Empty Payload",
			payload: []byte{},
			want:    0,
		},
		{
			name: "Test 2: SDP payload",
			payload: []byte(`v=0
			o=- 1423986 1423994 IN IP4 169.254.98.63
			s=AOIP44-serial-1614 : 2
			c=IN IP4 239.65.45.154/32
			t=0 0
			a=keywds:Dante
			m=audio 5004 RTP/AVP 97
			i=2 channels: TxChan 0, TxChan 1
			a=recvonly
			a=rtpmap:97 L24/48000/2
			a=ptime:1
			a=ts-refclk:ptp=IEEE1588-2008:00-00-00-FF-FE-00-00-00:0
			a=mediaclk:direct=142410716`),
			want: 49394,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeMsgIdHash(tc.payload)
			if tc.want != got {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
	}
}
