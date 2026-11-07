package transaction

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	prefixUint16 = 0xfd
	prefixUint32 = 0xfe
	prefixUint64 = 0xff
)

// DecodeCVarint는 r에서 CompactSize 바이트 배열을 읽고 uint64로 변환해서 반환한다.
func DecodeCVarint(r io.Reader) (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:1]); err != nil {
		return 0, fmt.Errorf("prefix 읽기 실패: %w", err)
	}
	prefix := b[0]

	var bodySize uint8
	var minBodyVal uint64
	switch prefix {
	case prefixUint16:
		bodySize, minBodyVal = 2, 0xfd
	case prefixUint32:
		bodySize, minBodyVal = 4, 0x10000
	case prefixUint64:
		bodySize, minBodyVal = 8, 0x100000000
	default:
		return uint64(b[0]), nil
	}
	if _, err := io.ReadFull(r, b[:bodySize]); err != nil {
		return 0, fmt.Errorf("body 읽기 실패: %w", err)
	}
	body := binary.LittleEndian.Uint64(b[:])
	if body < minBodyVal {
		return 0, fmt.Errorf("표준 CompactSize가 아님")
	}
	return body, nil
}

// EncodeCVarint는 n을 CompactSize 바이트 배열로 변환해서 반환한다.
func EncodeCVarint(n uint64) []byte {
	switch {
	case n <= 0xfc:
		return []byte{byte(n)}
	case n <= 0xffff:
		return binary.LittleEndian.AppendUint16([]byte{byte(0xfd)}, uint16(n))
	case n <= 0xffffffff:
		return binary.LittleEndian.AppendUint32([]byte{byte(0xfe)}, uint32(n))
	default:
		return binary.LittleEndian.AppendUint64([]byte{byte(0xff)}, uint64(n))
	}
}
