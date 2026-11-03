package transaction

import (
	"bytes"
	"encoding/binary"
	"io"
)

func DecodeCvarint(buf *bytes.Buffer) (uint64, error) {
	var (
		bPrefix []byte
		prefix  uint8
		bNum    []byte
		num     uint64
		n       int
		err     error
	)

	bPrefix = make([]byte, 1)
	if n, err = io.ReadFull(buf, bPrefix); n < len(bPrefix) || err != nil {
		return 0, nil
	} else {
		prefix = bPrefix[0]
	}

	switch prefix {
	case 0xfd:
		bNum = make([]byte, 2)
		if n, err = io.ReadFull(buf, bNum); n < len(bNum) || err != nil {
			return 0, nil
		} else {
			num = binary.LittleEndian.Uint64(bNum)
		}
	case 0xfe:
		bNum = make([]byte, 4)
		if n, err = io.ReadFull(buf, bNum); n < len(bNum) || err != nil {
			return 0, nil
		} else {
			num = binary.LittleEndian.Uint64(bNum)
		}
	case 0xff:
		bNum = make([]byte, 8)
		if n, err = io.ReadFull(buf, bNum); n < len(bNum) || err != nil {
			return 0, nil
		} else {
			num = binary.LittleEndian.Uint64(bNum)
		}
	default:
		num = uint64(prefix)
	}

	return num, nil
}

func EncodeCVarint(num uint64) ([]byte, error) {
	var (
		buf bytes.Buffer
		err error
	)

	switch {
	case num <= 0xfd-1:
		if err = binary.Write(&buf, binary.LittleEndian, uint8(num)); err != nil {
			return nil, err
		}
	case num <= 0x10000-1:
		buf.Write([]byte{uint8(0xfd)})
		if err = binary.Write(&buf, binary.LittleEndian, uint16(num)); err != nil {
			return nil, err
		}
	case num <= 0x100000000-1:
		buf.Write([]byte{uint8(0xfe)})
		if err = binary.Write(&buf, binary.LittleEndian, uint32(num)); err != nil {
			return nil, err
		}
	case num <= 0x10000000000000000-1:
		buf.Write([]byte{uint8(0xff)})
		if err = binary.Write(&buf, binary.LittleEndian, uint64(num)); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}
