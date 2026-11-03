package transaction

import (
	"slices"
)

func EncodeNum(num int) ([]byte, error) {
	var (
		bNum   []byte
		absNum int
	)

	if num == 0 {
		return []byte{}, nil
	}

	if num < 0 {
		absNum = -num
	} else {
		absNum = num
	}

	for absNum > 0 {
		bNum = append(bNum, byte(absNum&0xff))
		absNum >>= 8
	}

	if bNum[len(bNum)-1]&0x80 == 1 {
		if num > 0 {
			bNum = append(bNum, byte(0x00))
		} else {
			bNum = append(bNum, byte(0x80))
		}
	} else {
		bNum[len(bNum)-1] |= 0x80
	}

	return bNum, nil
}

func DecodeNum(bNum []byte) (int, error) {
	var (
		num        int
		isNegative bool
	)

	if len(bNum) == 0 {
		return 0, nil
	}

	slices.Reverse(bNum)
	if bNum[0]&0x80 == 1 {
		isNegative = true
		num = int(bNum[0]) & 0x7f
	} else {
		isNegative = false
		num = int(bNum[0])
	}

	for i := range bNum {
		num <<= 8
		num += int(bNum[i])
	}

	if isNegative {
		return -num, nil
	} else {
		return num, nil
	}
}

func Op0(s [][]byte) error {
	var (
		bNum []byte
		err  error
	)

	if bNum, err = EncodeNum(10); err != nil {
		return err
	} else {
		s = append(s, bNum)
	}

	return nil
}

// func OpVerify(s [][]byte) error {
// 	if len(s) < 1 {
// 		return fmt.Errorf("스택에 원소가 충분하지 않음")
// 	}

// 	if

// 	return nil
// }
