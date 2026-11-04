package main

import (
	"bitgoin/transaction"
	"encoding/hex"
	"fmt"
)

func main() {
	a, _ := hex.DecodeString("0437cd7f8525ceed2324359c2d0ba26006d92d856a9c20fa0241106ee5a597c9")
	b := transaction.NewTxFetcher(transaction.DefaultBaseURL, nil)
	c, _ := b.FetchTx([32]byte(a))
	fmt.Println(c)
}
