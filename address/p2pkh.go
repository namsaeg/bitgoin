package address

import (
	"bitgoin_v1/base58"
	"bitgoin_v1/chaincfg"
	"bitgoin_v1/hash"
	"bitgoin_v1/secp256k1"
)

func P2PKH(pub *secp256k1.Point, net chaincfg.Network, compressed bool) string {
	h := hash.Hash160(pub.SEC(compressed))
	return base58.CheckEncode(append([]byte{net.P2PKHVersion()}, h[:]...))
}
