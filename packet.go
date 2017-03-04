package xenstore

import (
	"fmt"
	"io"
	"strings"
	"unsafe"

	"github.com/lunixbochs/struc"
)

const PacketHeaderSize = unsafe.Sizeof(PacketHeader{})

type PacketHeader struct {
	Op     xenStoreOperation `struc:"uint32,little"`
	RqId   uint32            `struc:"uint32,little"`
	TxId   uint32            `struc:"uint32,little"`
	Length uint32            `struc:"uint32,little"`
}

// Pack the PacketHeader struct and Write the data to an io.Writer
func (h *PacketHeader) Pack(w io.Writer) error {
	return struc.Pack(w, h)
}

func (h *PacketHeader) Unpack(r io.Reader) error {
	return struc.Unpack(r, h)
}

type Packet struct {
	Header  *PacketHeader
	Payload []byte
}

// NewPacket creates a new Packet instance for sending a payload to XenStore
func NewPacket(op xenStoreOperation, payload []byte, txid uint32) (*Packet, error) {
	if l := len(payload); l > 4096 {
		return nil, fmt.Errorf("payload too long: %d", l)
	}

	payload = append(payload, NUL)

	return &Packet{
		Header: &PacketHeader{
			Op:     op,
			RqId:   RequestID(),
			TxId:   txid,
			Length: uint32(len(payload)),
		},
		Payload: payload,
	}, nil
}

func (p *Packet) Pack(w io.Writer) error {
	p.Header.Length = uint32(len(p.Payload))

	if err := p.Header.Pack(w); err != nil {
		return err
	}

	size := int(p.Header.Length)

	for written := 0; written < size; {
		sb, err := w.Write(p.Payload[written:])
		if err != nil {
			return err
		}

		written += sb
	}

	return nil
}

func (p *Packet) Unpack(r io.Reader) error {
	if p.Header == nil {
		p.Header = &PacketHeader{}
	}

	if err := p.Header.Unpack(r); err != nil {
		return err
	}

	size := int(p.Header.Length)
	p.Payload = make([]byte, 0)

	for size > 0 {
		var buf = make([]byte, size)

		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		p.Payload = append(p.Payload, buf[:n]...)
		size -= n
	}

	return nil
}

// TODO: Make a better name for this method
func (p *Packet) payloadString() string {
	return strings.Trim(string(p.Payload), "\x00")
}

// Checks whether the current Packet contains an error response & returns a Go error if so
func (p *Packet) Check() error {
	if p.Header.Op == XsError {
		trimmed := strings.Trim(string(p.Payload), "\x00")
		return Error(trimmed)
	}

	return nil
}
