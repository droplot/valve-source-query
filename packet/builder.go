package packet

import "bytes"

type Builder struct {
	bytes.Buffer
}

func (b *Builder) WriteCString(s string) {
	b.WriteString(s)
	b.WriteByte(0)
}

func (b *Builder) WriteBytes(bytes []byte) {
	b.Write(bytes)
}
