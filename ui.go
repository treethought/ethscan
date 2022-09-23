package main

import (
	"context"
	"os"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gdamore/tcell/v2"
	"github.com/kataras/golog"
)

type App struct {
	client   *ethclient.Client
	app      *cview.Application
	root     *cview.TabbedPanels
	focus    *cview.FocusManager
	bindings *cbind.Configuration
	broker   *Broker
	views    map[string]cview.Primitive
	log      *golog.Logger
	signer   types.Signer
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
		signer:   getSigner(context.TODO(), client),
	}
}

func (app *App) initBlockData() *cview.Flex {
	blockData := NewBlockData(app, nil)
	app.views["blockData"] = blockData

	wrap := cview.NewFlex()
	wrap.SetBackgroundTransparent(false)
	wrap.SetBackgroundColor(tcell.ColorDefault)
	wrap.SetDirection(cview.FlexRow)
	wrap.AddItem(blockData, 0, 1, true)
	return wrap
}

func (app *App) initBlockFeed() *cview.Flex {
	app.log.Debug("initializing block feed layout")
	blockFeed := NewBlockTable(app)
	app.views["blockFeed"] = blockFeed

	wrap := cview.NewFlex()
	wrap.SetBackgroundTransparent(false)
	wrap.SetBackgroundColor(tcell.ColorDefault)
	wrap.SetDirection(cview.FlexRow)
	wrap.AddItem(blockFeed, 0, 1, true)
	go blockFeed.watch(context.TODO())

	return wrap

}

func (app *App) initViews() {
	app.log.Debug("initializing views")

	blockFeed := app.initBlockFeed()
	blockData := app.initBlockData()

	dataPanels := cview.NewTabbedPanels()
	dataPanels.SetTitle("panels")
	dataPanels.AddTab("blockFeed", "blocks", blockFeed)
	dataPanels.AddTab("blockData", "block data", blockData)
	dataPanels.SetCurrentTab("blockFeed")
	dataPanels.SetBorder(false)
	dataPanels.SetPadding(0, 0, 0, 0)

	dataPanels.SetBackgroundColor(tcell.ColorDefault)
	dataPanels.SetTabBackgroundColor(tcell.ColorDefault)
	dataPanels.SetTabSwitcherDivider("", " | ", "")
	dataPanels.SetTabSwitcherAfterContent(true)

	app.root = dataPanels
	app.log.Debug("views ready")
}

func (a *App) setBindings() {
	a.bindings.SetKey(tcell.ModNone, tcell.KeyEsc, func(ev *tcell.EventKey) *tcell.EventKey {
		a.ShowBlocks()
		return nil
	})

	a.app.SetInputCapture(a.bindings.Capture)
}

func (a *App) ShowBlockData(b *types.Block) {
	a.log.Info("showing block data for: ", b.Hash().String())

	bd, ok := a.views["blockData"]
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

	a.root.SetCurrentTab("blockData")
	// a.app.SetRoot(bd, true)
}

func (a *App) ShowBlocks() {
	a.log.Info("showing blocks")
	a.root.SetCurrentTab("blockFeed")
}

func (a *App) Start() {
	defer a.app.HandlePanic()
	a.app.EnableMouse(true)

	a.setBindings()
	a.initViews()

	go a.broker.listenForBlocks(context.TODO())

	a.app.SetRoot(a.root, true)

	if err := a.app.Run(); err != nil {
		panic(err)
	}
}
