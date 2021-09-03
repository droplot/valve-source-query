package source

import (
	"errors"
	"github.com/icraftltd/valve-source-query/packet"
)

func (c *Client) getChallenge(header byte, fullResult byte) ([]byte, bool, error) {
	var b packet.Builder
	b.WriteBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, header, 0xFF, 0xFF, 0xFF, 0xFF})

	if err := c.send(b.Bytes()); err != nil {
		return nil, false, err
	}

	resp, err := c.receive()
	if err != nil {
		return nil, false, err
	}

	reader := packet.NewReader(resp)

	switch int32(reader.ReadUint32()) {
	case -2: // We received an unexpected full reply
		return resp, true, nil
	case -1: // Continue
	default:
		return nil, false, errors.New("source.Client.getChallenge: packet header mismatch")
	}

	switch reader.ReadUint8() {
	case 0x41: // Received a challenge number
		return resp[reader.Position() : reader.Position()+4], false, nil
	case fullResult: // Received full result
		return resp, true, nil
	}

	return nil, false, errors.New("source.Client.getChallenge: bad challenge response")
}
