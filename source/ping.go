package source

import (
	"errors"
	"github.com/icraftltd/valve-source-query/packet"
)

func (c *Client) Ping() (bool, error) {
	var b packet.Builder
	b.WriteBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x69})

	if err := c.send(b.Bytes()); err != nil {
		return false, err
	}

	response, err := c.receive()
	if err != nil {
		return false, err
	}

	reader := packet.NewReader(response)
	if reader.ReadInt32() != -1 {
		return false, errors.New("source.Client.Ping: packet header mismatch")
	}

	return reader.ReadUint8() == 0x6A, nil
}
