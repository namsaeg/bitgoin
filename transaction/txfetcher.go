package transaction

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultBaseURL = "https://mempool.space/api"

const maxTxHexSize = 8 << 20

var (
	ErrInvalidTxID  = errors.New("invalid txid")
	ErrTxIDMismatch = errors.New("fetched tx does not match requested txid")
)

type TxFetcher struct {
	baseURL string
	client  *http.Client
	cache   map[[32]byte]*Tx
}

func NewTxFetcher(baseURL string, client *http.Client) *TxFetcher {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &TxFetcher{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  client,
		cache:   make(map[[32]byte]*Tx),
	}
}

func (p *TxFetcher) Add(tx *Tx) {
	c := tx.Clone()
	p.cache[c.ID()] = c
}

func (p *TxFetcher) FetchTx(txID [32]byte) (*Tx, error) {
	if cached, ok := p.cache[txID]; ok {
		return cached.Clone(), nil
	}

	tx, err := p.download(txID)
	if err != nil {
		return nil, fmt.Errorf("fetch tx %x: %w", txID, err)
	}

	p.cache[txID] = tx
	return tx.Clone(), nil
}

func (p *TxFetcher) download(id [32]byte) (*Tx, error) {
	resp, err := p.client.Get(fmt.Sprintf("%s/tx/%x/hex", p.baseURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxTxHexSize+1))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if len(body) > maxTxHexSize {
		return nil, fmt.Errorf("response larger than %d bytes", maxTxHexSize)
	}

	raw, err := hex.DecodeString(strings.TrimSpace(string(body)))
	if err != nil {
		return nil, fmt.Errorf("decode response hex: %w", err)
	}

	r := bytes.NewReader(raw)
	tx, err := ParseTx(r)
	if err != nil {
		return nil, err
	}
	if r.Len() != 0 {
		return nil, fmt.Errorf("%d unexpected bytes after transaction", r.Len())
	}
	if got := tx.ID(); got != id {
		return nil, fmt.Errorf("%w: got %x", ErrTxIDMismatch, got)
	}
	return tx, nil
}
