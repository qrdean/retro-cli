package shared

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"

	"unsafe"
)

var VERSION byte = 1

const (
	AddStickyType byte = iota + 1
	VoteStickyType
	QuitType

	MaxPayloadSize = 10 << 20 // 10 MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload struct {
	Stringer   fmt.Stringer
	WriteTo    io.WriterTo
	ReaderFrom io.ReaderFrom
	Bytes      []byte
}

type AddStickyBytes []byte

type AddSticky struct {
	PosterId      uint32
	TopicId       uint32
	StickyMessage [StickyMessageSize]byte
}

func NewAddSticky(posterId, topicId uint32, msg string) (AddSticky, error) {
	var bytes [255]byte
	stringThing := []byte(msg)
	if len(stringThing) < 255 {
		copy(bytes[:len(stringThing)], stringThing)
	} else {
		return AddSticky{}, errors.New("msg too long")
	}

	return AddSticky{
		PosterId:      posterId,
		TopicId:       topicId,
		StickyMessage: bytes,
	}, nil
}

type VoteBytes []byte

type VoteSticky struct {
	TopicId uint32
}

type QuitBytes []byte

type Quit struct {
	ConnectionId uint32
}

func (m AddStickyBytes) Bytes() []byte  { return m }
func (m AddStickyBytes) String() string { return string(m) }

func (m AddStickyBytes) WriteTo(w io.Writer) (int64, error) {
	var bytesToWrite []byte
	bytesToWrite, err := binary.Append(bytesToWrite, binary.BigEndian, VERSION)
	// err := binary.Write(w, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	// err = binary.Write(w, binary.BigEndian, AddStickyType)
	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, AddStickyType)
	if err != nil {
		return n, err
	}
	n += 1

	// err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}

	n += 4

	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, m)

	o, err := w.Write(bytesToWrite)
	if err != nil {
		return n, err
	}

	return int64(o), err
}

func (m *AddStickyBytes) ReadFrom(r io.Reader) (int64, error) {
	var size uint32
	err := binary.Read(r, binary.BigEndian, &size)
	if err != nil {
		log.Printf("error reading size: %v\n", err)
		return 0, err
	}
	var n int64 = 4

	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	*m = make([]byte, size)
	o, err := r.Read(*m)
	if err != nil {
		return n, err
	}

	return n + int64(o), err
}

func (a AddSticky) MarshalBinary() []byte {
	datasize := int(unsafe.Sizeof(a))
	data := make([]byte, datasize)
	binary.BigEndian.PutUint32(data, a.PosterId)
	binary.BigEndian.PutUint32(data[4:], a.TopicId)
	copy(data[8:datasize], a.StickyMessage[:])
	return data
}

func (b AddStickyBytes) UnmarshalBinary() AddSticky {
	datasize := len(b)
	var addSticky AddSticky
	addSticky.PosterId = uint32(binary.BigEndian.Uint32(b))
	addSticky.TopicId = uint32(binary.BigEndian.Uint32(b[4:]))
	copy(addSticky.StickyMessage[:], b[8:datasize])

	return addSticky
}

func (m VoteBytes) Bytes() []byte  { return m }
func (m VoteBytes) String() string { return string(m) }

func (m VoteBytes) WriteTo(w io.Writer) (int64, error) {
	var bytesToWrite []byte
	// err := binary.Write(w, binary.BigEndian, VERSION)
	bytesToWrite, err := binary.Append(bytesToWrite, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	// err = binary.Write(w, binary.BigEndian, VoteStickyType)
	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, VoteStickyType)
	if err != nil {
		return n, err
	}
	n += 1

	// err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}

	n += 4

	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, m)
	o, err := w.Write(bytesToWrite)
	if err != nil {
		return n, err
	}

	return int64(o), err
}

func (m *VoteBytes) ReadFrom(r io.Reader) (int64, error) {
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
	if err != nil {
		return n, err
	}

	return n + int64(o), err
}

func (v VoteSticky) MarshalBinary() []byte {
	datasize := int(unsafe.Sizeof(v))
	data := make([]byte, datasize)
	binary.BigEndian.PutUint32(data, v.TopicId)
	return data
}

func (b VoteBytes) UnmarshalBinary() VoteSticky {
	var voteSticky VoteSticky
	voteSticky.TopicId = uint32(binary.BigEndian.Uint32(b))
	return voteSticky
}

func (m QuitBytes) Bytes() []byte  { return m }
func (m QuitBytes) String() string { return string(m) }

func (m QuitBytes) WriteTo(w io.Writer) (int64, error) {
	var bytesToWrite []byte
	// err := binary.Write(w, binary.BigEndian, VERSION)
	bytesToWrite, err := binary.Append(bytesToWrite, binary.BigEndian, VERSION)
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	// err = binary.Write(w, binary.BigEndian, QuitType)
	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, QuitType)
	if err != nil {
		return n, err
	}
	n += 1

	// err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}

	n += 4

	bytesToWrite, err = binary.Append(bytesToWrite, binary.BigEndian, m)
	o, err := w.Write(bytesToWrite)
	if err != nil {
		return n, err
	}

	return int64(o), err
}

func (m *QuitBytes) ReadFrom(r io.Reader) (int64, error) {
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
	if err != nil {
		return n, err
	}

	return n + int64(o), err
}

func (q Quit) MarshalBinary() []byte {
	datasize := int(unsafe.Sizeof(q))
	data := make([]byte, datasize)
	binary.BigEndian.PutUint32(data, q.ConnectionId)
	return data
}

func (b QuitBytes) UnmarshalBinary() Quit {
	var quit Quit
	quit.ConnectionId = uint32(binary.BigEndian.Uint32(b))
	return quit
}
