package sap

import (
	"errors"
)

var (
	errBufTooSmallForFlags       = errors.New("buffer too small for Flags")
	errBufTooSmallForAuthLength  = errors.New("buffer too small for Authentication Length")
	errBufTooSmallForAuthData    = errors.New("buffer too small for Authentication Data")
	errBufTooSmallForMsgIdHash   = errors.New("buffer too small for message id hash")
	errBufTooSmallForIPv4        = errors.New("buffer too small for IPv4 address")
	errBufTooSmallForIPv6        = errors.New("buffer too small for IPv6 address")
	errBufTooSmallForPayload     = errors.New("buffer too small for the payload")
	errBufTooSmallForPayloadType = errors.New("buffer too small for the payload type")
	errBufTooSmallForHeader      = errors.New("buffer too small for the header")
	errNoTrailingByteFound       = errors.New("didn't find the trailing byte from the buffer")
	errInvalidIPOnHeader         = errors.New("invalid IP in the OriginatingSource field on the Header Struct")
)
