package source

import (
	"errors"
	"github.com/icraftltd/valve-source-query/packet"
)

func (c *Client) Ping() (*PingResponse, error) {
	var b packet.Builder
	b.WriteBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x69})

	p := &PingResponse{Address: c.addr}
	if err := c.send(b.Bytes()); err != nil {
		return p, err
	}

	resp, err := c.receive()
	if err != nil {
		return p, err
	}

	reader := packet.NewReader(resp)
	if reader.ReadInt32() != -1 {
		return p, errors.New("source.Client.Ping: packet header mismatch")
	}

	p.Status = reader.ReadUint8() == 0x6A
	return p, nil
}
