package util

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Broker struct {
	sync.RWMutex
	client     *ethclient.Client
	blockSubs  []chan<- *types.Block
	txSubs     []chan<- *types.Transaction
	headerSubs []chan<- *types.Header
}

func NewBroker(client *ethclient.Client) *Broker {
	return &Broker{client: client}
}

func (b *Broker) SubscribeHeaders() chan *types.Header {
	b.Lock()
	defer b.Unlock()

	ch := make(chan *types.Header, 1)
	b.headerSubs = append(b.headerSubs, ch)
	return ch
}

func (b *Broker) SubcribeBlocks() chan *types.Block {
	b.Lock()
	defer b.Unlock()

	ch := make(chan *types.Block, 1)
	b.blockSubs = append(b.blockSubs, ch)
	return ch
}

func (b *Broker) SubscribeTransactions() chan *types.Transaction {
	b.Lock()
	defer b.Unlock()

	ch := make(chan *types.Transaction, 1)
	b.txSubs = append(b.txSubs, ch)
	return ch
}

func (b *Broker) pubHeader(header *types.Header) {
	b.RLock()
	defer b.RUnlock()
	for _, s := range b.headerSubs {
		s <- header
	}
}

func (b *Broker) pubBlock(block *types.Block) {
	b.RLock()
	defer b.RUnlock()
	for _, s := range b.blockSubs {
		s <- block
	}
}

func (b *Broker) pubTxn(t *types.Transaction) {
	b.RLock()
	defer b.RUnlock()
	for _, s := range b.txSubs {
		s <- t
	}
}

func (b *Broker) ListenForBlocks(ctx context.Context) {

	hChan := make(chan *types.Header)

	sub, err := b.client.SubscribeNewHead(ctx, hChan)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			fmt.Println(err)
			b.ListenForBlocks(ctx)
			return
		case h := <-hChan:
			b.pubHeader(h)
			block, err := b.client.BlockByHash(ctx, h.Hash())
			if err != nil {
				fmt.Println("failed to get block:", err)
				continue
			}

			b.pubBlock(block)
			for _, t := range block.Transactions() {
				b.pubTxn(t)
			}

		}
	}
}
