package transaction

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type TxFetcher struct {
	cache map[string]*Tx
}

func NewTxFetcher(cache map[string]*Tx) (*TxFetcher, error) {
	var txFetcher TxFetcher

	txFetcher.cache = make(map[string]*Tx, len(cache))
	for k := range cache {
		txFetcher.cache[k] = cache[k].Clone()
	}

	return &txFetcher, nil
}

func (p *TxFetcher) Cache() map[string]*Tx {
	var cache map[string]*Tx

	cache = make(map[string]*Tx, len(cache))
	for k := range p.cache {
		cache[k] = p.cache[k].Clone()
	}

	return cache
}
func (p *TxFetcher) SetCache(cache map[string]*Tx) error {
	p.cache = make(map[string]*Tx, len(cache))
	for k := range cache {
		p.cache[k] = cache[k].Clone()
	}

	return nil
}

func (p *TxFetcher) Clone() *TxFetcher {
	var txFetcher TxFetcher

	txFetcher.cache = make(map[string]*Tx, len(p.cache))
	for k := range p.cache {
		txFetcher.cache[k] = p.cache[k].Clone()
	}

	return &txFetcher
}

func (p *TxFetcher) String() string {
	var txFetcherStr strings.Builder

	for k, v := range p.cache {
		txFetcherStr.WriteString(k)
		txFetcherStr.WriteString("\n")
		txFetcherStr.WriteString(fmt.Sprintf("%v", v))
		txFetcherStr.WriteString("\n")
	}

	return fmt.Sprintf("TxFetcher(\n%s)", txFetcherStr.String())
}

func GetUrl() string {
	return "https://mempool.space/api"
}

func (p *TxFetcher) FetchTx(txId string) (*Tx, error) {
	var (
		val   *Tx
		ok    bool
		url   string
		resp  *http.Response
		body  []byte
		text  string
		bText []byte
		tx    *Tx
		err   error
	)

	val, ok = p.cache[txId]
	if ok {
		return val.Clone(), nil
	} else {
		url = fmt.Sprintf("%s/tx/%s/hex", GetUrl(), txId)

		if resp, err = http.Get(url); err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		if body, err = io.ReadAll(resp.Body); err != nil {
			return nil, err
		}

		text = strings.TrimSpace(string(body))

		if bText, err = hex.DecodeString(text); err != nil {
			return nil, err
		}

		if tx, err = ParseTx(bytes.NewBuffer(bText)); err != nil {
			return nil, err
		} else {
			p.cache[txId] = tx
			return p.cache[txId].Clone(), nil
		}
	}
}
