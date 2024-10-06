package shared

import (
	"errors"
	"fmt"
	"io"
)

var VERSION byte = 1

const (
	MaxPayloadSize = 10 << 20 // 10 MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload struct {
	Stringer   fmt.Stringer
	WriteTo    io.WriterTo
	ReaderFrom io.ReaderFrom
	Bytes      []byte
}


