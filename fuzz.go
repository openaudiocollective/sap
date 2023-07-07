//go:build gofuzz
// +build gofuzz

package sap

import "fmt"

func Fuzz(data []byte) int {
	predefined_errs := []error{
		errBufTooSmallForFlags,
		errBufTooSmallForAuthLength,
		errBufTooSmallForAuthData,
		errBufTooSmallForMsgIdHash,
		errBufTooSmallForIPv4,
		errBufTooSmallForIPv6,
		errBufTooSmallForPayload,
		errBufTooSmallForPayloadType,
		errBufTooSmallForHeader,
		errNoTrailingByteFound,
		errInvalidIPOnHeader,
	}

	p := Packet{}
	err := p.Unmarshal(data)

	if err != nil {
		// Now check if it's one of your defined errors
		for _, predeferr := range predefined_errs {
			if predeferr.Error() == err.Error() {
				fmt.Printf("The error '%s' is one of the predefined errors\n", err)
				return 1 // the input is invalid, no need to test it further
			}
		}
		return 0 // the input is invalid, no need to test it further
	}

	_, err = p.MarshalTo(data)
	if err != nil {
		// Now check if it's one of your defined errors
		for _, predeferr := range predefined_errs {
			if predeferr.Error() == err.Error() {
				fmt.Printf("The error '%s' is one of the predefined errors\n", err)
				return 1 // the input is invalid, no need to test it further
			}
		}
		return 0 // the input is invalid, no need to test it further
	}
	return 1 // the inpput was parsed successfully, so it's a good candidate for further testing
}
