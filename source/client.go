package source

import (
	"bytes"
	"compress/bzip2"
	"errors"
	"fmt"
	"github.com/icraftltd/valve-source-query/packet"
	"hash/crc32"
	"net"
	"strings"
	"time"
)

const (
	DefaultTimeout       = time.Second * 3
	DefaultPort          = 27015
	DefaultMaxPacketSize = 1400
)

var (
	ErrNilOption = errors.New("client.NewClient: invalid client option")
)

type GoldSourceMultiPacketHeader struct {
	ID uint32
	// The total number of packets in the response.
	Total uint8
	// The number of the packet. Starts at 0.
	Number uint8
	// Payload
	Payload    []byte
	Compressed bool
}
type MultiPacketHeader struct {
	// Size of the packet header
	Size int

	// Same as the Goldsource server meaning.
	// However, if the most significant bit is 1, then the response was compressed with bzip2 before being cut and sent.
	ID uint32

	// The total number of packets in the response.
	Total uint8

	// The number of the packet. Starts at 0.
	Number uint8

	/*
		(Orange Box Engine and above only.)
		Maximum size of packet before packet switching occurs.
		The default value is 1248 bytes (0x04E0), but the server administrator can decrease this.
		For older engine versions: the maximum and minimum size of the packet was unchangeable.
		AppIDs which are known not to contain this field: 215, 17550, 17700, and 240 when protocol = 7.
	*/
	SplitSize uint16

	// Indicates if payload is compressed w/bzip2
	Compressed bool

	// Payload
	Payload []byte
}

type Client struct {
	addr          string
	conn          net.Conn
	timeout       time.Duration
	maxPacketSize uint32
	buffer        []byte
	preOrange     bool
	appid         AppID
	wait          time.Duration
	next          time.Time
}

func TimeoutOption(timeout time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.timeout = timeout

		return nil
	}
}

func PreOrangeBox(pre bool) func(*Client) error {
	return func(c *Client) error {
		c.preOrange = pre

		return nil
	}
}

func SetAppID(appid int32) func(*Client) error {
	return func(c *Client) error {
		c.appid = AppID(appid)

		return nil
	}
}

// SetMaxPacketSize changes the maximum buffer size of a UDP packet
// Note that some games such as squad may use a non-standard packet size
// Refer to the game documentation to see if this needs to be changed
func SetMaxPacketSize(size uint32) func(*Client) error {
	return func(c *Client) error {
		c.maxPacketSize = size

		return nil
	}
}

func NewClient(addr string, options ...func(*Client) error) (c *Client, err error) {
	c = &Client{
		timeout:       DefaultTimeout,
		addr:          addr,
		maxPacketSize: DefaultMaxPacketSize,
	}

	for _, f := range options {
		if f == nil {
			return nil, ErrNilOption
		}
		if err = f(c); err != nil {
			return nil, err
		}
	}

	if !strings.Contains(c.addr, ":") {
		c.addr = fmt.Sprintf("%s:%d", c.addr, DefaultPort)
	}

	if c.conn, err = net.DialTimeout("udp", c.addr, c.timeout); err != nil {
		return nil, err
	}

	c.buffer = make([]byte, 0, c.maxPacketSize)

	return c, nil
}

func (c *Client) send(data []byte) error {
	c.enforceRateLimit()

	defer c.setNextQueryTime()

	if c.timeout > 0 {
		c.conn.SetWriteDeadline(c.extendedDeadline())
	}

	_, err := c.conn.Write(data)

	return err
}

func (c *Client) receive() ([]byte, error) {
	defer c.setNextQueryTime()

	if c.timeout > 0 {
		c.conn.SetReadDeadline(c.extendedDeadline())
	}

	size, err := c.conn.Read(c.buffer[0:c.maxPacketSize])

	if err != nil {
		return nil, err
	}

	buffer := make([]byte, size)

	copy(buffer, c.buffer[:size])

	return buffer, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) extendedDeadline() time.Time {
	return time.Now().Add(c.timeout)
}

func (c *Client) setNextQueryTime() {
	if c.wait != 0 {
		c.next = time.Now().Add(c.wait)
	}
}

func (c *Client) enforceRateLimit() {
	if c.wait == 0 {
		return
	}

	wait := c.next.Sub(time.Now())
	if wait > 0 {
		time.Sleep(wait)
	}
}

func (c *Client) parseMultiplePacketHeader(data []byte) (*GoldSourceMultiPacketHeader, error) {
	reader := packet.NewReader(data)

	if reader.ReadInt32() != -2 {
		return nil, errors.New("source.Client.parseMultiplePacketHeader: packet header mismatch")
	}

	header := &GoldSourceMultiPacketHeader{}
	header.ID = reader.ReadUint32()
	num := reader.ReadUint8()
	header.Total = num & 0xF
	header.Number = num >> 4
	header.Payload = data[reader.Position():]

	return header, nil
}

func (c *Client) collectMultiplePacketResponse(data []byte) ([]byte, error) {
	header, err := c.parseMultiplePacketHeader(data)

	if err != nil {
		return []byte{}, errors.New("source.Client.collectMultiplePacketResponse: packet header mismatch")
	}
	packets := make([]*GoldSourceMultiPacketHeader, header.Total)

	received := 0
	fullSize := 0

	for {
		if int(header.Number) >= len(packets) {
			return nil, errors.New("source.Client.collectMultiplePacketResponse: read out of bounds")
		}

		if packets[header.Number] != nil {
			return nil, errors.New("source.Client.collectMultiplePacketResponse: received same packet of same index")
		}

		packets[header.Number] = header

		fullSize += len(header.Payload)

		received++

		if received == len(packets) {
			break
		}

		data, err := c.receive()

		if err != nil {
			return nil, err
		}

		header, err = c.parseMultiplePacketHeader(data)

		if err != nil {
			return nil, err
		}
	}

	payload := make([]byte, fullSize)

	cursor := 0

	for _, header := range packets {
		copy(payload[cursor:cursor+len(header.Payload)], header.Payload)
		cursor += len(header.Payload)
	}

	// Includes decompressed size & crc32 sum as that is unread yet, so it's included as part of payload
	reader := packet.NewReader(payload)

	if packets[0].Compressed {
		decompressedSize := reader.ReadUint32()
		checkSum := reader.ReadUint32()

		if decompressedSize > uint32(1024*1024) {
			return nil, errors.New("source.Client.collectMultiplePacketResponse: bad bz2 decompression size")
		}

		decompressed := make([]byte, decompressedSize)

		bz2Reader := bzip2.NewReader(bytes.NewReader(data[reader.Position():]))

		n, err := bz2Reader.Read(decompressed)

		if err != nil {
			return nil, err
		}

		if n != int(decompressedSize) {
			return nil, errors.New("source.Client.collectMultiplePacketResponse: bad bz2 decompression size")
		}

		if crc32.ChecksumIEEE(decompressed) != checkSum {
			return nil, errors.New("source.Client.collectMultiplePacketResponse: bz2 decompressed checksum mismatches")
		}

		payload = decompressed
	}

	return payload, nil
}

func (c *Client) ResolvePlayersResponse(payloads []byte) (*Players, error) {
	reader := packet.NewReader(payloads)

	// Simple response now
	if reader.ReadInt32() != -1 {
		return nil, errors.New("source.Client.Players: packet header mismatch")
	}

	if reader.ReadUint8() != 0x44 {
		return nil, errors.New("source.Client.resolvePlayersResponse: bad players reply")
	}

	ps := &Players{Count: reader.ReadUint8()}
	for i := 0; i < int(ps.Count); i++ {
		p := &Player{
			Index:    reader.ReadUint8(),
			Name:     reader.ReadString(),
			Score:    reader.ReadUint32(),
			Duration: reader.ReadFloat32(),
		}

		if c.appid == 2400 {
			p.Ship = &ShipPlayer{Deaths: reader.ReadUint32(), Money: reader.ReadUint32()}
		}

		ps.Items = append(ps.Items, p)
	}

	return ps, nil
}
