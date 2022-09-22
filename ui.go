package main

import (
	"context"
	"os"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kataras/golog"
)

type App struct {
	client   *ethclient.Client
	app      *cview.Application
	focus    *cview.FocusManager
	bindings *cbind.Configuration
	broker   *Broker
	views    map[string]cview.Primitive
	log      *golog.Logger
}

func NewApp(client *ethclient.Client) *App {
	golog.SetLevel("debug")
	logFile, err := os.Create("./ethscan.log")
	if err != nil {
		panic(err)
	}
	log := golog.New().SetOutput(logFile)
	return &App{
		client:   client,
		app:      cview.NewApplication(),
		focus:    nil,
		bindings: cbind.NewConfiguration(),
		broker:   NewBroker(client),
		views:    make(map[string]cview.Primitive),
		log:      log,
	}
}

func (a *App) ShowBlockData(b *types.Block) {
	a.log.Info("showing block data for: ", b.Hash().String())
	bd, ok := a.views["block"]
	if !ok {
		a.log.Error("block view not set")
		panic("block view not set")
	}
	bdata, ok := bd.(*BlockData)
	if !ok {
		a.log.Error("was not blockdata")
		panic("was not blockdata")
	}
	a.log.Info("setting block")
	bdata.SetBlock(b)

	a.app.SetRoot(bd, true)
}

func (a *App) ShowBlocks() {
	a.log.Info("showing blocks")
	blocks := a.views["blocks"]
	a.app.SetRoot(blocks, true)
}

func (a *App) Start() {
	defer a.app.HandlePanic()

	blockTable := NewBlockTable(a)
	a.views["blocks"] = blockTable

	bd := NewBlockData(a, nil)
	a.views["block"] = bd

	go blockTable.watch(context.TODO())
	go a.broker.listenForBlocks(context.TODO())

	a.app.SetRoot(blockTable, true)

	if err := a.app.Run(); err != nil {
		panic(err)
	}
}
