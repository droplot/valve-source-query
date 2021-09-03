package packet

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
)

type Reader struct {
	data     []byte
	position int
}

func (r *Reader) Data() []byte {
	return r.data
}

func NewReader(data []byte) *Reader {
	return &Reader{data: data, position: 0}
}

func (r *Reader) CanRead(size int) bool {
	return r.position+size > len(r.data)
}

func (r *Reader) Position() int {
	return r.position
}

func (r *Reader) ReadIPv4() (net.IP, error) {
	if !r.CanRead(net.IPv4len) {
		return nil, errors.New("packet.Reader.ReadIPv4: read out of bounds")
	}

	ip := net.IP(r.data[r.Position() : r.position+net.IPv4len])
	r.position += net.IPv4len

	return ip, nil
}

func (r *Reader) ReadPort() (uint16, error) {
	if !r.CanRead(2) {
		return 0, errors.New("packet.Reader.ReadPort: read out of bounds")
	}

	p := binary.BigEndian.Uint16(r.data[r.position:])
	r.position += 2

	return p, nil
}

func (r *Reader) ReadUint8() uint8 {
	b := r.data[r.position]
	r.position++

	return b
}

func (r *Reader) ReadUint16() uint16 {
	u16 := binary.LittleEndian.Uint16(r.data[r.position:])
	r.position += 2

	return u16
}

func (r *Reader) ReadUint32() uint32 {
	u32 := binary.LittleEndian.Uint32(r.data[r.position:])
	r.position += 4
	return u32
}

func (r *Reader) ReadInt32() int32 {
	return int32(r.ReadUint32())
}

func (r *Reader) ReadUint64() uint64 {
	u64 := binary.LittleEndian.Uint64(r.data[r.position:])
	r.position += 8
	return u64
}

func (r *Reader) ReadFloat32() float32 {
	bits := r.ReadUint32()

	return math.Float32frombits(bits)
}

func (r *Reader) TryReadString() (string, bool) {
	start := r.position
	for r.position < len(r.data) {
		if r.data[r.position] == 0 {
			r.position++
			return string(r.data[start : r.position-1]), true
		}
		r.position++
	}
	return "", false
}

func (r *Reader) ReadString() string {
	start := r.position
	for {
		if r.data[r.position] == 0 {
			r.position++
			break
		}
		r.position++
	}
	return string(r.data[start : r.position-1])
}

func (r *Reader) More() bool {
	return r.position < len(r.data)
}
