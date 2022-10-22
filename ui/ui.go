package ui

import (
	"context"
	"log"
	"math/big"
	"os"
	"strconv"
	"sync"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gdamore/tcell/v2"
	"github.com/kataras/golog"
	"github.com/treethought/ethscan/util"
)

type State struct {
	sync.Mutex
	block           *types.Block
	txn             *types.Transaction
	contractAddress *common.Address
	history         []string
	currentView     string
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
func (s *State) SetContract(a *common.Address) {
	s.Lock()
	defer s.Unlock()
	s.contractAddress = a
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

type View interface {
	Update()
}

type App struct {
	client   *ethclient.Client
	app      *cview.Application
	root     *cview.TabbedPanels
	focus    *cview.FocusManager
	bindings *cbind.Configuration
	broker   *util.Broker
	views    map[string]View
	log      *golog.Logger
	signer   types.Signer
	State    *State
	config   *Config
}

func NewApp(config *Config) *App {
	golog.SetLevel("debug")

	client, err := ethclient.Dial(config.RpcUrl)
	if err != nil {
		log.Fatal(err)
	}

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
		broker:   util.NewBroker(client),
		views:    make(map[string]View),
		log:      log,
		signer:   util.GetSigner(context.TODO(), client),
		State:    &State{},
		config:   config,
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
	app.log.Debug("initializing txn data layout")

	txnData := NewTransactionData(app, nil)
	app.views["txnData"] = txnData

	txnLogs := NewTransactionLogs(app, nil)
	app.views["txnLogs"] = txnLogs

	wrap := cview.NewFlex()
	wrap.SetBackgroundTransparent(false)
	wrap.SetBackgroundColor(tcell.ColorDefault)
	wrap.SetDirection(cview.FlexRow)
	wrap.AddItem(txnData, 0, 1, true)
	wrap.AddItem(txnLogs, 0, 1, false)

	return wrap

}

func (app *App) initContractData() *cview.Flex {
	app.log.Debug("initializing contract data layout")

	contractData := NewContractForm(app, nil)
	app.views["contract"] = contractData

	wrap := cview.NewFlex()
	wrap.SetBackgroundTransparent(false)
	wrap.SetBackgroundColor(tcell.ColorDefault)
	wrap.SetDirection(cview.FlexRow)
	wrap.AddItem(contractData, 0, 1, true)

	return wrap

}

func (app *App) initViews() {
	app.log.Debug("initializing views")

	blockFeed := app.initBlockFeed()
	blockData := app.initBlockData()
	txnData := app.initTxnData()
	contractData := app.initContractData()

	dataPanels := cview.NewTabbedPanels()
	dataPanels.SetTitle("panels")
	dataPanels.AddTab("blockFeed", "blocks", blockFeed)
	dataPanels.AddTab("blockData", "block data", blockData)
	dataPanels.AddTab("txnData", "txn", txnData)
	dataPanels.AddTab("contract", "contract", contractData)
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
		a.State.RevertView()
		if a.State.currentView == "" {
			a.root.SetCurrentTab("blockFeed")
			return nil
		}
		a.root.SetCurrentTab(a.State.currentView)
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
	bd.Update()
	a.ShowView("blockData")
}

func (a *App) ShowView(v string) {
	a.root.SetCurrentTab(v)
	a.State.SetView(v)
}

func (a *App) ShowBlocks() {
	a.log.Info("showing blocks")
	a.ShowView("blockFeed")
}

func (a *App) ShowTransactonData(txn *types.Transaction) {
	a.log.Info("showing txn data for: ", txn.Hash().String())
	txnData := a.views["txnData"]
	txnLogs := a.views["txnLogs"]
	txnData.Update()
	txnLogs.Update()
	a.ShowView("txnData")
}
func (a *App) ShowContractData(txn *types.Transaction) {
	a.log.Info("showing txn data for: ", txn.Hash().String())
	contractData := a.views["contract"]
	contractData.Update()
	a.ShowView("contract")
}

func (a *App) Init() {
	a.app.EnableMouse(true)
	a.setBindings()
	a.initViews()
	a.app.SetRoot(a.root, true)
}

func (a *App) startWithBlockByNum(num int) error {
	big := big.NewInt(int64(num))
	block, err := a.client.BlockByNumber(context.TODO(), big)
	if err != nil {
		return err
	}
	a.State.SetBlock(block)
	a.State.SetTxn(nil)

	a.ShowBlockData(block)
	a.Start()
	return nil
}

func (a *App) startWithBlockByHash(hash string) error {
	h := common.HexToHash(hash)
	block, err := a.client.BlockByHash(context.TODO(), h)
	if err != nil {
		return err
	}
	a.State.SetBlock(block)
	a.ShowBlockData(block)
	a.Start()
	return nil

}

func (a *App) startWithTxn(hash string) error {
	ctx := context.TODO()
	h := common.HexToHash(hash)

	txn, _, err := a.client.TransactionByHash(context.Background(), h)
	if err != nil {
		return err
	}
	rec, err := a.client.TransactionReceipt(ctx, h)
	if err != nil {
		return err
	}
	block, err := a.client.BlockByHash(ctx, rec.BlockHash)
	if err != nil {
		return err
	}
	a.State.SetBlock(block)
	a.State.SetTxn(txn)
	a.ShowTransactonData(txn)
	a.Start()
	return nil

}

func (a *App) StartWith(ref string) {
	num, err := strconv.Atoi(ref)
	if err == nil {
		err = a.startWithBlockByNum(num)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = a.startWithBlockByHash(ref)
	if err == nil {
		return
	}

	err = a.startWithTxn(ref)
	if err == nil {
		return
	}

	log.Fatal("failed to get block or txn with hash: ", ref)

}

func (a *App) Start() {
	defer a.app.HandlePanic()

	go a.broker.ListenForBlocks(context.TODO())

	if err := a.app.Run(); err != nil {
		log.Fatal(err)
	}
}
