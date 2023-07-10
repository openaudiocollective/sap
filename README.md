# sap

Package sap provides SAP (Session Announcement Protocol) parsing compliant with [RFC 2974](https://datatracker.ietf.org/doc/html/rfc2974)

_This package is still experimental, breaking changes can happen._

## Usage

Get the Module
`go get github.com/openaudiocollective/sap`

Use the [NewPacket function](https://pkg.go.dev/github.com/openaudiocollective/sap#NewPacket) to create a Packet or create it manually. Unmarshal the packet to a byte slice or Marshal the packet into a structured Packet object.

*Check the [examples](./examples/) folder*

## Documentation

Head to the [documentation page](https://pkg.go.dev/github.com/openaudiocollective/sap) for more information.

## TODO

- Encryption/Decryption
- Compression
- Authentication

Inspired by [pion/sdp](https://github.com/pion/sdp) and [pion/rtp](https://github.com/pion/rtp). Thanks!
