package core

import (
	"encoding/binary"
)

type Frame struct {
	Flag    uint8
	Session uint16
	Size    uint16
	Payload []byte
}

func DecodeFrame(b []byte) Frame {
	bs := len(b)
	f := Frame{}
	f.Flag = b[0]
	f.Session = binary.BigEndian.Uint16(b[1:3])
	f.Size = binary.BigEndian.Uint16(b[3:5])
	f.Payload = nil
	if bs > 7 {
		f.Payload = b[5:]
	}
	return f
}

func (f Frame) Encode() []byte {
	var plen int

	if f.Payload != nil {
		plen = len(f.Payload)
	} else {
		plen = 0
	}

	totalSize := 1 + 2 + 2 + plen

	buf := make([]byte, totalSize)

	buf[0] = f.Flag
	binary.BigEndian.PutUint16(buf[1:3], f.Session)
	binary.BigEndian.PutUint16(buf[3:5], f.Size)

	if plen > 0 {
		copy(buf[5:], f.Payload)
	}

	return buf
}
