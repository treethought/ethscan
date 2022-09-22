package main

import (
	"fmt"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
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
	bd.SetBackgroundColor(tcell.ColorPink)

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
	d.app.log.Info("SHOWING BLOCKS")

	d.app.ShowBlocks()
	return _ev
}

func (d *BlockData) SetBlock(block *types.Block) {
	d.app.app.QueueUpdateDraw(func() {
		d.block = block
		d.render()
	})
}

func (d *BlockData) blockHeaders() *cview.List {
	l := cview.NewList()
	l.SetTitle("Headers")
	l.SetSelectedTextColor(tcell.ColorPink)

	number := d.block.Number().String()
	hash := d.block.Hash().String()
	time := formatUnixTime(d.block.Time())
	parent := d.block.ParentHash().String()
	coinbase := d.block.Coinbase().String()
	gasLimit := d.block.GasLimit()
	gasUsed := d.block.GasUsed()
	baseFee := d.block.BaseFee()
	root := d.block.Root().String()
	extraData := string(d.block.Extra())

	l.AddItem(cview.NewListItem(fmt.Sprintf("Number: %s", number)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Hash: %s", hash)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Time: %s", time)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Parent: %s", parent)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Coinbase (Proposer): %s", coinbase)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("GasLimit: %d", gasLimit)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("GasUsed: %d", gasUsed)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("BaseFee: %s", baseFee)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Root: %s", root)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("ExtraData: %s", extraData)))
	return l

}

func (d *BlockData) render() {
	d.Clear()

	if d.block == nil {
		return
	}

	d.SetBorders(true)
	d.SetRows(0, 0, 0)
	d.SetColumns(-1, -3, 0)
	d.SetBorders(true)

	d.AddItem(d.blockHeaders(), 0, 0, 1, 3, 0, 0, false)

	txns := NewTransactionTable(d.app, d.block)
	d.AddItem(txns, 1, 0, 2, 3, 0, 0, true)
}
