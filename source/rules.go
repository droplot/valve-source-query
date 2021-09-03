package source

import (
	"encoding/binary"
	"errors"
	"github.com/icraftltd/valve-source-query/packet"
)

func (c *Client) Rules() (*Rules, error) {
	resp, immediate, err := c.getChallenge(0x56, 0x45)
	if err != nil {
		return nil, err
	}

	if !immediate {
		pr := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x56, resp[0], resp[1], resp[2], resp[3]}
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
		return parseRulesInfo(resp)
	case -2:
		resp, err = c.collectMultiplePacketResponse(resp)

		if err != nil {
			return nil, err
		}

		return parseRulesInfo(resp)
	}

	return nil, errors.New("source.Client.Rules: packet header mismatch")
}

func parseRulesInfo(data []byte) (*Rules, error) {
	reader := packet.NewReader(data)

	// Simple response now

	if reader.ReadInt32() != -1 {
		return nil, errors.New("source.Client.Rules.parseRulesInfo: packet header mismatch")
	}

	if reader.ReadUint8() != 0x45 {
		return nil, errors.New("source.Client.Rules.parseRulesInfo: bad rules reply")
	}

	rules := &Rules{Count: reader.ReadUint16()}
	rules.Items = make(map[string]string, rules.Count)
	for i := 0; i < int(rules.Count); i++ {
		key, ok := reader.TryReadString()

		if !ok {
			break
		}

		val, ok := reader.TryReadString()

		if !ok {
			break
		}

		rules.Items[key] = val
	}

	return rules, nil
}
