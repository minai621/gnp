package ch04

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType
	MaxPayloadSize uint32 = 10 << 20 // 10MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

type Binary []byte

func (b Binary) Bytes() []byte {
	return b
}

func (b Binary) String() string {
	return string(b)
}

func (b Binary) WriterTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, BinaryType) // 1 byte
	if err != nil {
		return 0, err
	}
	var n int64

	err = binary.Write(w, binary.BigEndian, uint32(len(b))) // 4 bytes
	if err != nil {
		return n, err
	}

	n += 4

	o, err := w.Write(b) // 페이로드 
	n += int64(o)
	return n, err
}
