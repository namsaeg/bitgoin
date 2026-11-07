package txfetcher

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultBaseUrl = "https://mempool.space/api"
	maxTxHexSize   = 8 << 20
)

type TxFetcher struct {
	baseUrl string
	client  *http.Client
	cache   map[[32]byte]*Tx
}

func NewTxFetcher(baseUrl string, client *http.Client) *TxFetcher {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	return &TxFetcher{
		baseUrl: strings.TrimSuffix(baseUrl, "/"),
		client:  client,
		cache:   make(map[[32]byte]*Tx),
	}
}

func (p *TxFetcher) FetchTx(txId [32]byte) (*Tx, error) {
	if cached, ok := p.cache[txId]; ok {
		return cached.Clone(), nil
	}
	tx, err := p.download(txId)
	if err != nil {
		return nil, fmt.Errorf("트랜잭션(%x) 다운로드 실패: %w", txId, err)
	}
	p.cache[txId] = tx
	return tx.Clone(), nil
}

func (p *TxFetcher) download(id [32]byte) (*Tx, error) {
	resp, err := p.client.Get(fmt.Sprintf("%s/tx/%x/hex", p.baseUrl, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("잘못된 http 스테이터스: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxTxHexSize+1))
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}
	if len(body) > maxTxHexSize {
		return nil, fmt.Errorf("응답의 크기가 %d 바이트보다 큼", maxTxHexSize)
	}

	raw, err := hex.DecodeString(strings.TrimSpace(string(body)))
	if err != nil {
		return nil, fmt.Errorf("16진수 응답 해독하기 실패: %w", err)
	}

	r := bytes.NewReader(raw)
	tx, err := ParseTx(r)
	if err != nil {
		return nil, err
	}
	if r.Len() != 0 {
		return nil, fmt.Errorf("트랜잭션을 파싱하고 바이트가 남음 %d", r.Len())
	}
	if got := tx.Id(); got != id {
		return nil, fmt.Errorf("요청한 id와 가져온 트랜잭션의 id가 다름")
	}
	return tx, nil
}
