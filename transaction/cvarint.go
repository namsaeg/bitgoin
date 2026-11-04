package transaction

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	prefixUint16 = 0xfd
	prefixUint32 = 0xfe
	prefixUint64 = 0xff
)

var ErrNonCanonicalCVarint = errors.New("non-canonical CompactSize varint")

func DecodeCVarint(r io.Reader) (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:1]); err != nil {
		return 0, fmt.Errorf("read CompactSize varint prefix: %w", err)
	}

	var size int
	var minVal uint64

	switch b[0] {
	case prefixUint16:
		size, minVal = 2, prefixUint16
	case prefixUint32:
		size, minVal = 4, 0x10000
	case prefixUint64:
		size, minVal = 8, 0x100000000
	default:
		return uint64(b[0]), nil
	}

	if _, err := io.ReadFull(r, b[:size]); err != nil {
		return 0, fmt.Errorf("read CompactSize varint body (%d bytes): %w", size, err)
	}

	v := binary.LittleEndian.Uint64(b[:])
	if v < minVal {
		return 0, ErrNonCanonicalCVarint
	}
	return v, nil
}

func EncodeCVarint(v uint64) []byte {
	switch {
	case v < prefixUint16:
		return []byte{byte(v)}
	case v <= 0xffff:
		return binary.LittleEndian.AppendUint16([]byte{prefixUint16}, uint16(v))
	case v <= 0xffffffff:
		return binary.LittleEndian.AppendUint32([]byte{prefixUint32}, uint32(v))
	default:
		return binary.LittleEndian.AppendUint64([]byte{prefixUint64}, v)
	}
}
