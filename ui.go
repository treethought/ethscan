package main

import (
	"context"
	"os"
	"sync"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gdamore/tcell/v2"
	"github.com/kataras/golog"
)

type State struct {
	sync.Mutex
	block       *types.Block
	txn         *types.Transaction
	history     []string
	currentView string
}

func (s *State) SetBlock(b *types.Block) {
	s.Lock()
	defer s.Unlock()
	s.block = b
}
func (s *State) SetTxn(t *types.Transaction) {
	s.Lock()
	defer s.Unlock()
	s.txn = t
}

func (s *State) SetView(p string) {
	s.Lock()
	defer s.Unlock()
	s.history = append([]string{s.currentView}, s.history...)
	s.currentView = p
}

func (s *State) RevertView() {
	s.Lock()
	defer s.Unlock()
	if len(s.history) == 0 {
		return
	}
	prev := s.history[0]
	if len(s.history) == 1 {
		s.history = []string{}
		s.currentView = prev
		return
	}
	s.history = s.history[1:]
	s.currentView = prev
	return
}
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
	state    *State
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
		state:    &State{},
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

func (app *App) initTxnData() *cview.Flex {
	app.log.Debug("initializing block feed layout")
	txnData := NewTransactionData(app, nil)
	app.views["txnData"] = txnData

	wrap := cview.NewFlex()
	wrap.SetBackgroundTransparent(false)
	wrap.SetBackgroundColor(tcell.ColorDefault)
	wrap.SetDirection(cview.FlexRow)
	wrap.AddItem(txnData, 0, 1, true)

	return wrap

}

func (app *App) initViews() {
	app.log.Debug("initializing views")

	blockFeed := app.initBlockFeed()
	blockData := app.initBlockData()
	txnData := app.initTxnData()

	dataPanels := cview.NewTabbedPanels()
	dataPanels.SetTitle("panels")
	dataPanels.AddTab("blockFeed", "blocks", blockFeed)
	dataPanels.AddTab("blockData", "block data", blockData)
	dataPanels.AddTab("txnData", "txn", txnData)
	dataPanels.SetCurrentTab("blockFeed")
	dataPanels.SetBorder(false)
	dataPanels.SetPadding(0, 0, 0, 0)

	dataPanels.SetBackgroundColor(tcell.ColorDefault)
	dataPanels.SetTabBackgroundColor(tcell.ColorDefault)
	dataPanels.SetTabSwitcherDivider("", " | ", "")
	dataPanels.SetTabSwitcherAfterContent(true)

	app.root = dataPanels
	app.log.Debug("views ready")
	app.ShowView("blockFeed")
}

func (a *App) setBindings() {
	a.bindings.SetKey(tcell.ModNone, tcell.KeyEsc, func(ev *tcell.EventKey) *tcell.EventKey {
		a.state.RevertView()
		if a.state.currentView == "" {
			a.root.SetCurrentTab("blockFeed")
			return nil
		}
		a.root.SetCurrentTab(a.state.currentView)
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

	a.ShowView("blockData")
}

func (a *App) ShowView(v string) {
	a.root.SetCurrentTab(v)
	a.state.SetView(v)
}

func (a *App) ShowBlocks() {
	a.log.Info("showing blocks")
	a.ShowView("blockFeed")
}

func (a *App) ShowTransactonData(txn *types.Transaction) {
	a.log.Info("showing txn data for: ", txn.Hash().String())

	// TODO: setup state to pull current tx from inside widget
	// instead of calling SetTransaction
	bd, ok := a.views["txnData"]
	if !ok {
		a.log.Error("txn view not set")
		panic("txn view not set")
	}
	tdata, ok := bd.(*TransactionData)
	if !ok {
		a.log.Error("was not txn data")
		panic("was not txn data")
	}
	a.log.Info("setting txn")

	tdata.SetTransaction(txn)

	a.ShowView("txnData")
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
