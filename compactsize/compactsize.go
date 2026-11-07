package compactsize

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	ErrNonCanonical = errors.New("compactsize: 최소 인코딩이 아님")
	ErrTooLarge     = errors.New("compactsize: 값이 상한을 넘음")
)

const (
	prefixUint16 = 0xfd
	prefixUint32 = 0xfe
	prefixUint64 = 0xff
)

func Read(r io.Reader) (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:1]); err != nil {
		return 0, fmt.Errorf("compactsize: 접두 읽기: %w", err)
	}

	prefix := b[0]
	bs, minVal := bodySize(prefix)
	if bs == 0 {
		return uint64(prefix), nil
	}
	if _, err := io.ReadFull(r, b[:bs]); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return 0, fmt.Errorf("compactsize: %d바이트 본문 읽기: %w", bs, err)
	}

	n := binary.LittleEndian.Uint64(b[:])
	if n < minVal {
		return 0, fmt.Errorf("%w: 접두 %#x에 값 %d", ErrNonCanonical, prefix, n)
	}
	return n, nil
}

func ReadMax(r io.Reader, max uint64) (uint64, error) {
	n, err := Read(r)
	if err != nil {
		return 0, err
	}
	if n > max {
		return 0, fmt.Errorf("%w: %d > %d", ErrTooLarge, n, max)
	}
	return n, nil
}

func Decode(b []byte) (n uint64, size int, err error) {
	if len(b) == 0 {
		return 0, 0, fmt.Errorf("compactsize: 접두 읽기: %w", io.EOF)
	}
	prefix := b[0]
	bs, minVal := bodySize(prefix)
	if bs == 0 {
		return uint64(prefix), 1, nil
	}
	if len(b)-1 < bs {
		return 0, 0, fmt.Errorf("compactsize: %d바이트 본문 읽기: %w", bs, io.ErrUnexpectedEOF)
	}
	var buf [8]byte
	copy(buf[:], b[1:1+bs])
	n = binary.LittleEndian.Uint64(buf[:])
	if n < minVal {
		return 0, 0, fmt.Errorf("%w: 접두 %#x에 값 %d", ErrNonCanonical, prefix, n)
	}
	return n, 1 + bs, nil
}

func Size(n uint64) int {
	switch {
	case n < 0xfd:
		return 1
	case n <= 0xffff:
		return 3
	case n <= 0xffffffff:
		return 5
	default:
		return 9
	}
}

func bodySize(prefix byte) (size int, minVal uint64) {
	switch prefix {
	case prefixUint16:
		return 2, 0xfd
	case prefixUint32:
		return 4, 0x10000
	case prefixUint64:
		return 8, 0x100000000
	default:
		return 0, 0
	}
}

func Append(b []byte, n uint64) []byte {
	switch {
	case n <= 0xfc:
		return append(b, byte(n))
	case n <= 0xffff:
		return binary.LittleEndian.AppendUint16(append(b, prefixUint16), uint16(n))
	case n <= 0xffffffff:
		return binary.LittleEndian.AppendUint32(append(b, prefixUint32), uint32(n))
	default:
		return binary.LittleEndian.AppendUint64(append(b, prefixUint64), uint64(n))
	}
}

func Encode(n uint64) []byte { return Append(nil, n) }
