package main

import (
	"bitgoin_v1/address"
	"bitgoin_v1/chaincfg"
	"bitgoin_v1/secp256k1/ecdsa"
	"bitgoin_v1/wif"
	"fmt"
	"math/big"
)

func main() {
	a, _ := ecdsa.NewPrivateKey(big.NewInt(69740947))
	b := address.P2PKH(a.PublicKey(), chaincfg.Testnet, true)
	fmt.Println(b)
	c := wif.Encode(a, chaincfg.Testnet, true)
	fmt.Println(c)
}
