package main

import (
	"context"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type App struct {
	client   *ethclient.Client
	app      *cview.Application
	focus    *cview.FocusManager
	bindings *cbind.Configuration
	broker   *Broker
}

func NewApp(client *ethclient.Client) App {
	return App{
		client:   client,
		app:      cview.NewApplication(),
		focus:    nil,
		bindings: cbind.NewConfiguration(),
		broker:   NewBroker(client),
	}
}

func (a App) ShowBlockData(b *types.Block) {
	bd := NewBlockData(a.client, b)
	a.app.SetRoot(bd, true)
}

func (a App) Start() {
	blockTable := NewBlockTable(a)
	blockTable.SetBorder(true)

	go blockTable.watch(context.TODO())
	go a.broker.listenForBlocks(context.TODO())

	a.app.SetRoot(blockTable, true)

	if err := a.app.Run(); err != nil {
		panic(err)
	}
}
