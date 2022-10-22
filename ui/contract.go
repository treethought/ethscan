package ui

import (
	"fmt"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gdamore/tcell/v2"
	"github.com/treethought/ethscan/util"
)

type ContractForm struct {
	*cview.TabbedPanels
	address  *common.Address
	contract *util.ContractData
	app      *App
	bindings *cbind.Configuration
}

func NewContractForm(app *App, address *common.Address) *ContractForm {
	c := &ContractForm{
		TabbedPanels: cview.NewTabbedPanels(),
		contract:     nil,
		address:      address,
		app:          app,
	}

	c.initBindings()
	return c

}

func (c *ContractForm) initBindings() {
	c.bindings = cbind.NewConfiguration()
	c.SetInputCapture(c.bindings.Capture)
	c.bindings.SetRune(tcell.ModNone, 's', c.showSource)
	c.bindings.SetRune(tcell.ModNone, 'o', c.showSource)
	c.bindings.SetKey(tcell.ModNone, tcell.KeyTab, c.toggleTab)

}
func (c *ContractForm) toggleTab(ev *tcell.EventKey) *tcell.EventKey {
	tabs := []string{"overview", "source"}
	nextIdx := 0
	for i, t := range tabs {
		if t == c.GetCurrentTab() {
			nextIdx = i + 1
		}
	}
	if nextIdx > len(tabs)-1 {
		nextIdx = 0
	}

	c.SetCurrentTab(tabs[nextIdx])
	return nil
}

func (c *ContractForm) showSource(ev *tcell.EventKey) *tcell.EventKey {
	c.SetCurrentTab("source")
	return nil
}
func (c *ContractForm) showOverview(ev *tcell.EventKey) *tcell.EventKey {
	c.SetCurrentTab("o")
	return nil
}

func (c *ContractForm) Update() {
	c.TabbedPanels = cview.NewTabbedPanels()

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

type ABISummary struct {
	cview.TreeView
}

func (c *ContractForm) render() {

	grid := cview.NewGrid()
	grid.SetTitle("Contract")
	grid.SetBorder(true)
	grid.SetBorders(true)

	grid.SetRows(0, 2)
	grid.SetColumns(0, 0)

	if c.contract == nil {
		return
	}

	info := cview.NewList()
	info.SetTitle("Info")
	// info.SetBorder(true)
	info.AddItem(cview.NewListItem(fmt.Sprintf("Name: %s", c.contract.ContractName)))
	info.AddItem(cview.NewListItem(fmt.Sprintf("Compiler: %s", c.contract.CompilerVersion)))

	settings := cview.NewList()
	settings.SetTitle("Settings")

	opt := fmt.Sprintf("OptimizationUsed: %t with %d runs", c.contract.OptimizationUsed, c.contract.Runs)

	settings.AddItem(cview.NewListItem(opt))
	settings.AddItem(cview.NewListItem(fmt.Sprintf("Proxy: %t", c.contract.Proxy)))
	settings.AddItem(cview.NewListItem(fmt.Sprintf("EVM Verison: %s", c.contract.EVMVersion)))
	settings.AddItem(cview.NewListItem(fmt.Sprintf("LicenseType: %s", c.contract.LicenseType)))

	cabi := c.contract.ABI

	abi := cview.NewList()
	abi.SetTitle("ABI Summary")
	abi.AddItem(cview.NewListItem(fmt.Sprint(cabi.Constructor.String())))

	for _, m := range c.contract.ABI.Methods {
		abi.AddItem(cview.NewListItem(m.String()))
	}

	grid.AddItem(info, 0, 0, 1, 1, 0, 0, true)
	grid.AddItem(settings, 0, 1, 1, 1, 0, 0, true)
	grid.AddItem(abi, 1, 0, 2, 2, 0, 0, true)

	source := cview.NewTextView()
	source.SetTitle("Source Code")
	source.SetText(c.contract.SourceCode)
	source.SetScrollable(true)

	c.AddTab("overview", "Overview", grid)
	c.AddTab("source", "Source Code", source)

	c.SetBorder(false)
	c.SetPadding(0, 0, 0, 0)

	c.SetBackgroundColor(tcell.ColorDefault)
	c.SetTabBackgroundColor(tcell.ColorDefault)
	c.SetTabSwitcherDivider("", " | ", "")
	c.SetTabSwitcherAfterContent(false)
	c.SetCurrentTab("overview")
	c.app.app.SetFocus(c)

}
