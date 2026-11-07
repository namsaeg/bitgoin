package chaincfg

import "fmt"

type Network byte

const (
	Mainnet Network = iota + 1
	Testnet
)

func (n Network) String() string {
	switch n {
	case Mainnet:
		return "mainnet"
	case Testnet:
		return "testnet"
	default:
		return fmt.Sprintf("Network(%d)", int(n))
	}
}

func (n Network) P2PKHVersion() byte {
	switch n {
	case Mainnet:
		return 0x00
	case Testnet:
		return 0x6f
	}
	panic("chaincfg: 알 수 없는 네트워크 " + n.String())
}

func (n Network) WIFVersion() byte {
	switch n {
	case Mainnet:
		return 0x80
	case Testnet:
		return 0xef
	}
	panic("chaincfg: 알 수 없는 네트워크 " + n.String())
}
