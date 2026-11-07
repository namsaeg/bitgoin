package hash

import (
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

func Hash160(b []byte) [20]byte {
	s := sha256.Sum256(b)
	h := ripemd160.New()
	h.Write(s[:])
	var out [20]byte
	copy(out[:], h.Sum(nil))
	return out
}

func Hash256(b []byte) [32]byte {
	h := sha256.Sum256(b)
	h = sha256.Sum256(h[:])
	return h
}
