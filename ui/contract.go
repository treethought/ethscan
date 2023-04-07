package ui

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gdamore/tcell/v2"
	"github.com/treethought/ethscan/util"
)

type ContractForm struct {
	*cview.Grid
	address    *common.Address
	contract   *util.ContractData
	app        *App
	bindings   *cbind.Configuration
	sourceView *SourceCodeView
}

func NewContractForm(app *App, address *common.Address) *ContractForm {
	c := &ContractForm{
		Grid:     cview.NewGrid(),
		contract: nil,
		address:  address,
		app:      app,
	}

	c.initBindings()
	return c

}

func (c *ContractForm) initBindings() {
	c.bindings = cbind.NewConfiguration()
	c.SetInputCapture(c.bindings.Capture)
	c.bindings.SetRune(tcell.ModNone, 's', c.showSource)
}

func (c *ContractForm) showSource(ev *tcell.EventKey) *tcell.EventKey {
	c.app.app.SetRoot(c.sourceView, true)
	return nil
}

func (c *ContractForm) Update() {

	c.address = c.app.State.contractAddress
	if c.address == nil {
		c.app.log.Error("no contract address set")
		c.app.app.QueueUpdateDraw(c.render)
		return
	}

	contract, err := util.GetContractData(c.address.String(), c.app.config.EtherscanKey)
	if err != nil {
		c.app.log.Error("failed to get contract data: %v", err)
		return
	}
	c.contract = contract

	c.app.app.QueueUpdateDraw(func() {
		c.render()
	})

}

func contractInfo(c *ContractForm) *cview.List {
	info := cview.NewList()
	info.SetTitle("Info")
	// info.SetBorder(true)
	info.AddItem(cview.NewListItem(fmt.Sprintf("Name: %s", c.contract.ContractName)))
	info.AddItem(cview.NewListItem(fmt.Sprintf("Compiler: %s", c.contract.CompilerVersion)))
	info.AddItem(cview.NewListItem(fmt.Sprintf("Implementation: %s", c.contract.Implementation)))

	info.SetInputCapture(c.bindings.Capture)
	return info
}

func contracSettings(c *ContractForm) *cview.List {
	settings := cview.NewList()
	settings.SetTitle("Settings")

	opt := fmt.Sprintf("OptimizationUsed: %t with %d runs", c.contract.OptimizationUsed, c.contract.Runs)

	settings.AddItem(cview.NewListItem(opt))
	settings.AddItem(cview.NewListItem(fmt.Sprintf("Proxy: %t", c.contract.Proxy)))
	settings.AddItem(cview.NewListItem(fmt.Sprintf("EVM Verison: %s", c.contract.EVMVersion)))
	settings.AddItem(cview.NewListItem(fmt.Sprintf("LicenseType: %s", c.contract.LicenseType)))

	settings.SetInputCapture(c.bindings.Capture)
	return settings
}

func contractABI(c *ContractForm) *cview.List {
	cabi := c.contract.ABI

	abiInfo := cview.NewList()
	abiInfo.SetTitle("ABI (hit `s` to view source code)")
	abiInfo.SetBorder(true)
	abiInfo.AddItem(cview.NewListItem(fmt.Sprint(cabi.Constructor.String())))

	for _, m := range c.contract.ABI.Methods {
		abiInfo.AddItem(cview.NewListItem(m.String()))
	}

	abiInfo.SetInputCapture(c.bindings.Capture)
	return abiInfo
}

func contractConstructorArgs(c *ContractForm) *cview.List {
	consArgs := cview.NewList()
	consArgs.SetBorder(true)
	consArgs.SetTitle("Constructor Arguments")

	consArgs.AddItem(cview.NewListItem(c.contract.ConstructorArguments))

	argBytes, err := hex.DecodeString(c.contract.ConstructorArguments)
	if err != nil {
		c.app.log.Error("failed to decode args: ", err)
	}
	d, err := c.contract.ABI.Constructor.Inputs.Unpack(argBytes)
	if err != nil {
		c.app.log.Error("failed to unpack args: ", err)
	}

	for i, d := range d {
		arg := c.contract.ABI.Constructor.Inputs[i]
		txt := fmt.Sprintf("[%d] %s (%s) %s", i, arg.Name, arg.Type, d)
		consArgs.AddItem(cview.NewListItem(txt))
	}
	return consArgs
}

func (c *ContractForm) render() {
	c.Clear()
	c.SetTitle("Contract")
	c.SetBorder(true)
	c.SetBorders(true)

	c.SetRows(0, 0, 0, 1)
	c.SetColumns(0, 0)

	if c.contract == nil {
		return
	}

	info := contractInfo(c)
	settings := contracSettings(c)

	abiInfo := contractABI(c)
	consArgs := contractConstructorArgs(c)

	c.AddItem(info, 0, 0, 1, 1, 0, 0, true)
	c.AddItem(settings, 0, 1, 1, 1, 0, 0, true)
	c.AddItem(consArgs, 1, 0, 1, 2, 0, 0, true)
	c.AddItem(abiInfo, 2, 0, 2, 2, 0, 0, true)

	source := NewSourceCodeView(c.app, c.contract)
	c.sourceView = source

	c.SetBackgroundTransparent(false)
	c.SetBackgroundColor(tcell.ColorDefault)

}

type SourceCodeView struct {
	*cview.TextView
	app      *App
	bindings *cbind.Configuration
	contract *util.ContractData
}

func NewSourceCodeView(app *App, contract *util.ContractData) *SourceCodeView {
	c := &SourceCodeView{
		TextView: cview.NewTextView(),
		app:      app,
		contract: contract,
	}

	if c.contract != nil {
		c.SetTitle(c.contract.ContractName)
		c.SetBorder(true)
		c.SetText(c.contract.SourceCode)
	}

	c.initBindings()
	c.SetDoneFunc(func(_ tcell.Key) {
		c.app.app.SetRoot(c.app.root, true)
	})
	return c
}

func (s *SourceCodeView) initBindings() {
	s.bindings = cbind.NewConfiguration()
	s.bindings.SetKey(tcell.ModNone, tcell.KeyEsc, s.onDone)
	s.bindings.SetRune(tcell.ModNone, 'e', s.openEditor)
	s.SetInputCapture(s.bindings.Capture)
}

func (s *SourceCodeView) onDone(ev *tcell.EventKey) *tcell.EventKey {
	s.app.app.SetRoot(s.app.root, true)
	return nil
}

func (s *SourceCodeView) openEditor(ev *tcell.EventKey) *tcell.EventKey {
	if s.contract == nil {
		return nil
	}
	tmp, err := os.CreateTemp("", "ethscan*.sol")
	if err != nil {
		s.app.log.Error("failed to create tmp file: ", err)
		return nil
	}
	_, err = tmp.Write([]byte(s.contract.SourceCode))
	if err != nil {
		s.app.log.Error("failed to write contract: ", err)
		return nil
	}
	tmp.Close()

	s.app.app.Suspend(func() {
		editor := os.Getenv("EDITOR")
		path, err := exec.LookPath(editor)
		if err != nil {
			s.app.log.Error("failed to lookup $EDITOR")
		}

		cmd := exec.Command(path, tmp.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			s.app.log.Error("failed to open editor: ", err)
			return
		}
	})
	return nil
}
