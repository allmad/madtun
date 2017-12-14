package madtun_api

import "encoding/binary"

type Buffer struct {
	buffer []byte
	bw     binary.ByteOrder
	offset int
}

func NewBuffer(buf []byte) *Buffer {
	return &Buffer{
		buffer: buf,
		bw:     binary.BigEndian,
	}
}

func (b *Buffer) WriteUint32(n uint32) {
	b.bw.PutUint32(b.buffer[b.offset:], n)
	b.offset += 4
}

func (b *Buffer) Write(buf []byte) {
	b.offset += copy(b.buffer[b.offset:], buf)
}

func (b *Buffer) WriteUint16(n uint16) {
	b.bw.PutUint16(b.buffer[b.offset:], n)
	b.offset += 2
}

func (b *Buffer) Len() int {
	return b.offset
}

func (b *Buffer) Cap() int {
	return len(b.buffer)
}

func (b *Buffer) Avail() int {
	return len(b.buffer) - b.offset
}

func (b *Buffer) Reset() {
	b.offset = 0
}

func (b *Buffer) Bytes() []byte {
	return b.buffer[:b.offset]
}
