package source

import (
	"encoding/binary"
	"errors"
)

func (c *Client) Players() (*Players, error) {
	resp, immediate, err := c.getChallenge(0x55, 0x44)
	if err != nil {
		return nil, err
	}

	if !immediate {
		pr := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x55, resp[0], resp[1], resp[2], resp[3]}
		if err := c.send(pr); err != nil {
			return nil, err
		}

		resp, err = c.receive()
		if err != nil {
			return nil, err
		}
	}

	// Read header (long 4 bytes)
	switch int32(binary.LittleEndian.Uint32(resp)) {
	case -1:
		return c.ResolvePlayersResponse(resp)
	case -2:
		resp, err = c.collectMultiplePacketResponse(resp)

		if err != nil {
			return nil, err
		}

		return c.ResolvePlayersResponse(resp)
	}

	return nil, errors.New("source.Client.Players: packet header mismatch")
}
