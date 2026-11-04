package transaction

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"slices"
	"strings"
)

var (
	ErrSegwitUnsupported    = errors.New("segwit transactions are not supported")
	ErrTxOutIndexOutOfRange = errors.New("output index out of range")
	ErrCoinbaseFee          = errors.New("coinbase transaction has no fee")
	ErrNegativeFee          = errors.New("outputs exceed inputs")
	ErrAmountOverflow       = errors.New("amount sum overflows uint64")
)

const maxPrealloc = 1024

type Tx struct {
	version  uint32
	txIns    []*TxIn
	txOuts   []*TxOut
	locktime uint32
}

func NewTx(version uint32, txIns []*TxIn, txOuts []*TxOut, locktime uint32) *Tx {
	return &Tx{
		version:  version,
		txIns:    cloneAll(txIns),
		txOuts:   cloneAll(txOuts),
		locktime: locktime,
	}
}

func (p *Tx) Version() uint32 {
	return p.version
}

func (p *Tx) SetVersion(version uint32) {
	p.version = version
}

func (p *Tx) TxIns() []*TxIn {
	return cloneAll(p.txIns)
}

func (p *Tx) SetTxIns(txIns []*TxIn) {
	p.txIns = cloneAll(txIns)
}

func (p *Tx) TxOuts() []*TxOut {
	return cloneAll(p.txOuts)
}

func (p *Tx) SetTxOuts(txOuts []*TxOut) {
	p.txOuts = cloneAll(txOuts)
}

func (p *Tx) TxOut(i uint32) (*TxOut, error) {
	if i >= uint32(len(p.txOuts)) {
		return nil, fmt.Errorf("txOuts index out of range: index %d (len %d)", i, len(p.txOuts))
	}
	return p.txOuts[i], nil
}

func (p *Tx) Locktime() uint32 {
	return p.locktime
}

func (p *Tx) SetLocktime(locktime uint32) {
	p.locktime = locktime
}

func (p *Tx) Clone() *Tx {
	return &Tx{
		version:  p.version,
		txIns:    cloneAll(p.txIns),
		txOuts:   cloneAll(p.txOuts),
		locktime: p.locktime,
	}
}

func (p *Tx) ID() [32]byte {
	first := sha256.Sum256(p.Serialize())
	id := sha256.Sum256(first[:])
	slices.Reverse(id[:])
	return id
}

func (p *Tx) IsCoinbase() bool {
	return len(p.txIns) == 1 &&
		p.txIns[0].prevTxID == [32]byte{} &&
		p.txIns[0].prevTxOutIdx == 0xffffffff
}

func (p *Tx) Fee(f TxFetcher) (uint64, error) {
	if p.IsCoinbase() {
		return 0, ErrCoinbaseFee
	}
	var in, out uint64
	for i, txIn := range p.txIns {
		v, err := txIn.Value(f)
		if err != nil {
			return 0, fmt.Errorf("txin %d: %w", i, err)
		}
		if in, err = addAmount(in, v); err != nil {
			return 0, err
		}
	}
	for _, txOut := range p.txOuts {
		var err error
		if out, err = addAmount(out, txOut.Amount()); err != nil {
			return 0, err
		}
	}
	if out > in {
		return 0, fmt.Errorf("%w: in %d sat, out %d sat", ErrNegativeFee, in, out)
	}
	return in - out, nil
}

func (p *Tx) GetMessage(f *TxFetcher) ([]byte, error) {
	tmpTx := p.Clone()
	for i, txIn := range tmpTx.txIns {
		prevTx, err := f.FetchTx(txIn.PrevTxID())
		if err != nil {
			return nil, fmt.Errorf("cannot get prev tx: %w", err)
		}
		prevTxOut, err := prevTx.TxOut(txIn.prevTxOutIdx)
		if err != nil {
			return nil, fmt.Errorf("cannot get prev tx out: %w", err)
		}
		tmpTx.txIns[i].SetScriptSig(prevTxOut.scriptPubKey)
	}
	return tmpTx.Serialize(), nil
}

func addAmount(a, b uint64) (uint64, error) {
	sum, carry := bits.Add64(a, b, 0)
	if carry != 0 {
		return 0, ErrAmountOverflow
	}
	return sum, nil
}

func (p *Tx) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Tx(\nversion: %d\ntxIns:\n", p.version)
	for _, in := range p.txIns {
		sb.WriteString(in.String())
		sb.WriteByte('\n')
	}
	sb.WriteString("txOuts:\n")
	for _, out := range p.txOuts {
		sb.WriteString(out.String())
		sb.WriteByte('\n')
	}
	fmt.Fprintf(&sb, "locktime: %d\n)", p.locktime)
	return sb.String()
}

func (p *Tx) Serialize() []byte {
	b := binary.LittleEndian.AppendUint32(nil, p.version)
	b = append(b, EncodeCVarint(uint64(len(p.txIns)))...)
	for _, in := range p.txIns {
		b = append(b, in.Serialize()...)
	}
	b = append(b, EncodeCVarint(uint64(len(p.txOuts)))...)
	for _, out := range p.txOuts {
		b = append(b, out.Serialize()...)
	}
	return binary.LittleEndian.AppendUint32(b, p.locktime)
}

func ParseTx(r io.Reader) (*Tx, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("tx version: %w", err)
	}
	tx := &Tx{version: binary.LittleEndian.Uint32(b[:])}

	nIn, err := DecodeCVarint(r)
	if err != nil {
		return nil, fmt.Errorf("tx input count: %w", noEOF(err))
	}
	if nIn == 0 {
		return nil, ErrSegwitUnsupported
	}
	if tx.txIns, err = parseList(r, nIn, ParseTxIn, "txin"); err != nil {
		return nil, err
	}

	nOut, err := DecodeCVarint(r)
	if err != nil {
		return nil, fmt.Errorf("tx output count: %w", noEOF(err))
	}
	if tx.txOuts, err = parseList(r, nOut, ParseTxOut, "txout"); err != nil {
		return nil, err
	}

	if _, err := io.ReadFull(r, b[:]); err != nil {
		return nil, fmt.Errorf("tx locktime: %w", noEOF(err))
	}
	tx.locktime = binary.LittleEndian.Uint32(b[:])

	return tx, nil
}

func parseList[T any](r io.Reader, n uint64, parse func(io.Reader) (T, error), what string) ([]T, error) {
	items := make([]T, 0, min(n, maxPrealloc))
	for i := uint64(0); i < n; i++ {
		item, err := parse(r)
		if err != nil {
			return nil, fmt.Errorf("%s %d: %w", what, i, noEOF(err))
		}
		items = append(items, item)
	}
	return items, nil
}

func cloneAll[T interface{ Clone() T }](src []T) []T {
	dst := make([]T, len(src))
	for i, v := range src {
		dst[i] = v.Clone()
	}
	return dst
}
