package madtun_api

import (
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/chzyer/logex"
)

const packetMinSize = 4 + 2 + 4
const maxPayloadSize = 1 << 20

var ErrPayloadTooLarge = logex.Define("payload is too large: %v")

type Packet struct {
	ReqId   uint32
	Type    Type
	Payload []byte
}

func (p *Packet) Decode(r io.Reader, buf []byte) error {
	tmp := buf
	_, err := io.ReadAtLeast(r, tmp[:packetMinSize], packetMinSize)
	if err != nil {
		return logex.Trace(err)
	}

	p.ReqId = binary.BigEndian.Uint32(tmp)
	tmp = tmp[4:]
	p.Type = Type(binary.BigEndian.Uint16(tmp))
	tmp = tmp[2:]
	payloadLength := int(binary.BigEndian.Uint32(tmp))
	tmp = tmp[4:]
	if payloadLength > maxPayloadSize {
		return ErrPayloadTooLarge.Format(payloadLength)
	}
	tmp = buf
	if _, err := io.ReadAtLeast(r, tmp, payloadLength); err != nil {
		return logex.Trace(err)
	}
	p.Payload = make([]byte, payloadLength)
	copy(p.Payload, tmp[:payloadLength])
	return nil
}

func (p *Packet) Size() int {
	required := int(unsafe.Sizeof(p.ReqId)) + int(unsafe.Sizeof(p.Type))
	required += 2 + len(p.Payload)
	return required
}

func (p *Packet) Marshal(buf *Buffer) {
	if buf.Avail() < p.Size() {
		panic("must check size before Marshal")
	}

	buf.WriteUint32(p.ReqId)
	buf.WriteUint16(uint16(p.Type))
	buf.WriteUint32(uint32(len(p.Payload)))
	buf.Write(p.Payload)
}
