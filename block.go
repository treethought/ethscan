package main

import (
	"fmt"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/aquilax/truncate"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gdamore/tcell/v2"
)

type BlockData struct {
	*cview.Grid
	block    *types.Block
	bindings *cbind.Configuration
	app      *App
}

func NewBlockData(app *App, block *types.Block) *BlockData {
	bd := &BlockData{app: app, block: block, Grid: cview.NewGrid()}

	bd.setBindings()
	bd.render()
	return bd
}

func (d *BlockData) setBindings() {
	bind := cbind.NewConfiguration()
	bind.SetKey(tcell.ModNone, tcell.KeyEsc, d.focusBlocks)
	d.bindings = bind
	d.SetInputCapture(d.bindings.Capture)
}

func (d *BlockData) focusBlocks(_ev *tcell.EventKey) *tcell.EventKey {
	d.app.ShowBlocks()
	return _ev
}

func (d *BlockData) SetBlock(block *types.Block) {
	d.app.app.QueueUpdateDraw(func() {
		if d.block != nil && d.block.Hash() == block.Hash() {
			return
		}
		d.block = block
		d.render()
	})
}
func (d *BlockData) Update() {
	curBlock := d.app.state.block
	d.SetBlock(curBlock)
}

func (d *BlockData) blockHeaders() *cview.Flex {
	f := cview.NewFlex()
	f.SetDirection(cview.FlexColumn)

	basic := cview.NewList()
	basic.SetTitle("basic")
	basic.SetBorder(true)

	details := cview.NewList()
	details.SetTitle("details")
	details.SetBorder(true)

	number := d.block.Number().String()
	hash := truncate.Truncate(d.block.Hash().String(), 10, "...", truncate.PositionMiddle)
	time := formatUnixTime(d.block.Time())
	parent := truncate.Truncate(d.block.ParentHash().String(), 20, "...", truncate.PositionMiddle)
	coinbase := truncate.Truncate(d.block.Coinbase().String(), 10, "...", truncate.PositionMiddle)
	gasLimit := d.block.GasLimit()
	gasUsed := d.block.GasUsed()
	baseFee := d.block.BaseFee().String()
	root := truncate.Truncate(d.block.Root().String(), 10, "...", truncate.PositionMiddle)
	extraData := string(d.block.Extra())

	basic.AddItem(cview.NewListItem(fmt.Sprintf("Number: %s", number)))
	basic.AddItem(cview.NewListItem(fmt.Sprintf("Hash: %s", hash)))
	basic.AddItem(cview.NewListItem(fmt.Sprintf("Time: %s", time)))
	basic.AddItem(cview.NewListItem(fmt.Sprintf("Parent: %s", parent)))

	details.AddItem(cview.NewListItem(fmt.Sprintf("Coinbase (Proposer): %s", coinbase)))
	details.AddItem(cview.NewListItem(fmt.Sprintf("GasLimit: %d", gasLimit)))
	details.AddItem(cview.NewListItem(fmt.Sprintf("GasUsed: %d", gasUsed)))
	details.AddItem(cview.NewListItem(fmt.Sprintf("BaseFee: %s", baseFee)))
	details.AddItem(cview.NewListItem(fmt.Sprintf("Root: %s", root)))
	details.AddItem(cview.NewListItem(fmt.Sprintf("ExtraData: %s", extraData)))

	f.AddItem(basic, 0, 1, true)
	f.AddItem(details, 0, 1, false)
	return f

}

func (d *BlockData) render() {
	d.Clear()

	if d.block == nil {
		return
	}

	d.SetTitle(fmt.Sprintf("Block #%d", d.block.Number()))
	d.SetRows(0, 0, 0)
	d.SetColumns(-1, -3, 0)

	d.AddItem(d.blockHeaders(), 0, 0, 1, 3, 0, 0, false)

	txns := NewTransactionTable(d.app, d.block)
	d.AddItem(txns, 1, 0, 2, 3, 0, 0, true)

	d.app.app.SetFocus(txns)

}
