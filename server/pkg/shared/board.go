package shared

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

const (
	PointerType byte = iota + 1
	TopicType
	StickyType
)

type PointerBytes []byte

func (m PointerBytes) Bytes() []byte  { return m }
func (m PointerBytes) String() string { return string(m) }

func (m PointerBytes) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, PointerType)
	if err != nil {
		return n, err
	}
	n += 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}

	n += 4

	o, err := w.Write(m)

	return n + int64(o), err
}

func (m *PointerBytes) ReadFrom(r io.Reader) (int64, error) {
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

type TopicBytes []byte

func (m TopicBytes) Bytes() []byte  { return m }
func (m TopicBytes) String() string { return string(m) }

func (m TopicBytes) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, TopicType)
	if err != nil {
		return n, err
	}
	n += 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}

	n += 4

	o, err := w.Write(m)

	return n + int64(o), err
}

func (m *TopicBytes) ReadFrom(r io.Reader) (int64, error) {
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

type StickyBytes []byte

func (m StickyBytes) Bytes() []byte  { return m }
func (m StickyBytes) String() string { return string(m) }

func (m StickyBytes) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, StickyType)
	if err != nil {
		return n, err
	}
	n += 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}

	n += 4

	o, err := w.Write(m)

	return n + int64(o), err
}

func (m *StickyBytes) ReadFrom(r io.Reader) (int64, error) {
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

type Pointer struct {
	PointerId uint32
}

const HeaderSize = 255

type Topic struct {
	Id     uint32
	Header [HeaderSize]byte
}

const StickyMessageSize = 255

type Sticky struct {
	Id            uint32
	PosterId      uint32
	TopicId       uint32
	Votes         uint32
	StickyMessage [StickyMessageSize]byte
}

func (s Sticky) MarshalBinary() []byte {
	datasize := int(unsafe.Sizeof(s))
	data := make([]byte, datasize)
	binary.BigEndian.PutUint32(data, s.Id)
	binary.BigEndian.PutUint32(data[4:], s.PosterId)
	binary.BigEndian.PutUint32(data[8:], s.TopicId)
	binary.BigEndian.PutUint32(data[12:], s.Votes)
	copy(data[16:datasize], s.StickyMessage[:])
	return data
}

func (s StickyBytes) UnmarshalBinaryStick() Sticky {
	datasize := len(s)
	// fmt.Printf("datasize %v\n", datasize)
	var sticky Sticky
	sticky.Id = uint32(binary.BigEndian.Uint32(s))
	sticky.PosterId = uint32(binary.BigEndian.Uint32(s[4:]))
	sticky.TopicId = uint32(binary.BigEndian.Uint32(s[8:]))
	sticky.Votes = uint32(binary.BigEndian.Uint32(s[12:]))
	fmt.Printf("msg %v\n", string(s[16:datasize]))
	copy(sticky.StickyMessage[:], s[16:datasize])
	return sticky
}

func (t Topic) MarshalBinary() []byte {
	datasize := int(unsafe.Sizeof(t))
	data := make([]byte, datasize)
	binary.BigEndian.PutUint32(data, t.Id)
	copy(data[4:HeaderSize], t.Header[:])

	return data
}

func (t TopicBytes) UnmarshalTopic() Topic {
	var topic Topic
	topic.Id = uint32(binary.BigEndian.Uint32(t))
	copy(topic.Header[:], t[4:HeaderSize])

	return topic
}

func (p Pointer) MarshalBinary() []byte {
	datasize := int(unsafe.Sizeof(p))
	data := make([]byte, datasize)
	binary.BigEndian.PutUint32(data, p.PointerId)
	return data
}

func (p PointerBytes) UnmarshalPointer() Pointer {
	var pointer Pointer
	pointer.PointerId = uint32(binary.BigEndian.Uint32(p))
	return pointer
}
