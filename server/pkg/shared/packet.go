package shared

import (
	"encoding/binary"
	"io"
)

type Byte []byte

func (m Byte) Bytes() []byte  { return m }
func (m Byte) String() string { return string(m) }

type Packet struct {
	Type byte
	Byte Byte
}

func (m Packet) Bytes() []byte  { return m.Byte }
func (m Packet) String() string { return string(m.Byte) }

func (m Packet) WriteTo(w io.Writer) (int64, error) {
	var bytesToWrite []byte
	bytesToWrite, err := binary.Append(bytesToWrite, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, m.Type)
	if err != nil {
		return n, err
	}
	n += 1

	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, uint32(len(m.Byte)))
	if err != nil {
		return n, err
	}

	n += 4

	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, m.Byte)
	o, err := w.Write(bytesToWrite)
	if err != nil {
		return n, err
	}

	return int64(o), err
}

func (m *Byte) ReadFrom(r io.Reader) (int64, error) {
	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		return 0, err
	}
	var n int64 = 4

	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	*m = make([]byte, size)
	o, err := r.Read(*m)

	return n + int64(o), err
}

func MarshalBinaryTopicLength(d uint32) []byte {
	data := make([]byte, 10)
	binary.BigEndian.PutUint32(data, d)
	return data
}

func UnmarshalPointerTopicLength(b []byte) uint32 {
	return uint32(binary.BigEndian.Uint32(b))
}
