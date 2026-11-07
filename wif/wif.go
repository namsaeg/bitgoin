package wif

import (
	"bitgoin_v1/base58"
	"bitgoin_v1/chaincfg"
	"bitgoin_v1/secp256k1/ecdsa"
)

func Encode(key *ecdsa.PrivateKey, net chaincfg.Network, compressed bool) string {
	d := key.Bytes()
	b := make([]byte, 0, 34)
	b = append(b, net.WIFVersion())
	b = append(b, d[:]...)
	if compressed {
		b = append(b, 0x01)
	}
	return base58.CheckEncode(b)
}
