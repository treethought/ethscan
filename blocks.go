package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/aquilax/truncate"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gdamore/tcell/v2"
)

const truncSize = 12

type blockRow struct {
	header *types.Header
}

func newBlockRow(h *types.Header) blockRow {
	return blockRow{header: h}
}

func (r blockRow) Time() *cview.TableCell {
	human := formatUnixTime(r.header.Time)
	return cview.NewTableCell(human)
}

func (r blockRow) Number() *cview.TableCell {
	return cview.NewTableCell(r.header.Number.String())
}

func (r blockRow) Hash() *cview.TableCell {
	hash := r.header.Hash().String()
	tr := truncate.Truncate(hash, truncSize, "...", truncate.PositionMiddle)
	cell := cview.NewTableCell(tr)
	// set the row's reference
	cell.SetReference(r.header.Number)
	return cell
}

func (r blockRow) Parent() *cview.TableCell {
	hash := r.header.ParentHash.String()
	tr := truncate.Truncate(hash, truncSize, "...", truncate.PositionMiddle)
	return cview.NewTableCell(tr)
}

func (r blockRow) Miner() *cview.TableCell {
	addr := r.header.Coinbase.String()
	tr := truncate.Truncate(addr, truncSize, "...", truncate.PositionMiddle)
	return cview.NewTableCell(tr)
}

func (r blockRow) GasLimit() *cview.TableCell {
	val := fmt.Sprint(r.header.GasLimit)
	tr := truncate.Truncate(val, truncSize, "...", truncate.PositionMiddle)
	return cview.NewTableCell(tr)
}
func (r blockRow) GasUsed() *cview.TableCell {
	val := fmt.Sprint(r.header.GasUsed)
	tr := truncate.Truncate(val, truncSize, "...", truncate.PositionMiddle)
	return cview.NewTableCell(tr)
}
func (r blockRow) BaseFee() *cview.TableCell {
	val := r.header.BaseFee.String()
	tr := truncate.Truncate(val, truncSize, "...", truncate.PositionMiddle)
	return cview.NewTableCell(tr)
}
func (r blockRow) StateRoot() *cview.TableCell {
	val := r.header.Root.String()
	tr := truncate.Truncate(val, truncSize, "...", truncate.PositionMiddle)
	return cview.NewTableCell(tr)
}
func (r blockRow) TxnLength() *cview.TableCell {
	val := r.header.Root.String()
	tr := truncate.Truncate(val, truncSize, "...", truncate.PositionMiddle)
	return cview.NewTableCell(tr)
}
func (r blockRow) ExtraData() *cview.TableCell {
	val := string(r.header.Extra)
	// tr := truncate.Truncate(val, truncSize, "...", truncate.PositionEnd)
	return cview.NewTableCell(val)
}

type BlockTable struct {
	*cview.Table
	app      *App
	headers  []*types.Header
	fields   []string
	bindings *cbind.Configuration

	ch chan *types.Header
}

func NewBlockTable(app *App) *BlockTable {
	table := &BlockTable{
		Table:   cview.NewTable(),
		headers: nil,
		ch:      make(chan *types.Header),
		app:     app,
	}
	table.fields = []string{
		"Time",
		"Number",
		"Hash",
		"Miner",
		"GasLimit",
		"GasUsed",
		"BaseFee",
		"ExtraData",
	}

	table.SetBorders(true)
	table.SetFixed(0, 0)
	table.SetSelectable(true, false)
	table.SetSelectedStyle(tcell.ColorBlueViolet, tcell.ColorDefault, 0)
	table.SetSelectedFunc(func(row, _c int) {
		if row == 0 {
			return
		}
		table.app.log.Info("Selected block - row ", row)

		// referene is currently only set on hash cell = col 2
		cell := table.GetCell(row, 2)
		ref := cell.GetReference()
		num, ok := ref.(*big.Int)
		if !ok {
			log.Fatal("reference was not a big.Int blockNumber")
		}
		table.app.log.Info("Row reference number: ", num.String())
		block, err := app.client.BlockByNumber(context.TODO(), num)
		if err != nil {
			table.app.log.Fatal(err)
		}
		table.app.state.SetBlock(block)
		table.app.state.SetTxn(nil)
		table.app.ShowBlockData(block)

	})

	table.initBindings()

	table.setTableHeader()
	table.ch = app.broker.SubscribeHeaders()
	return table

}
func (t *BlockTable) Update() {}

func (t *BlockTable) initBindings() {
	t.bindings = cbind.NewConfiguration()
	t.SetInputCapture(t.bindings.Capture)
	t.bindings.SetRune(tcell.ModNone, 'o', t.handleOpen)

}

func (t *BlockTable) getCurrentRef() *big.Int {

	row, _ := t.GetSelection()
	t.app.log.Info("getting current selected block, row: ", row)
	ref := t.GetCell(row, 2).GetReference()
	num, ok := ref.(*big.Int)
	if !ok {
		t.app.log.Error("failed to get block ref")
		return nil
	}
	return num
}

func (t *BlockTable) handleOpen(ev *tcell.EventKey) *tcell.EventKey {
	curNum := t.getCurrentRef()
	if curNum == nil {
		return nil
	}
	url := fmt.Sprintf("https://etherscan.io/block/%s", curNum.String())
	openbrowser(url)
	return nil

}

func (t BlockTable) setTableHeader() {
	for col, field := range t.fields {
		cell := cview.NewTableCell(field)
		cell.SetSelectable(false)
		t.SetCell(0, col, cell)
	}
}

func (t BlockTable) addHeaderRow(r blockRow) {
	row := t.GetRowCount()
	t.SetCell(row, 0, r.Time())
	t.SetCell(row, 1, r.Number())
	t.SetCell(row, 2, r.Hash())
	t.SetCell(row, 3, r.Miner())
	t.SetCell(row, 4, r.GasLimit())
	t.SetCell(row, 5, r.GasUsed())
	t.SetCell(row, 6, r.BaseFee())
	t.SetCell(row, 7, r.ExtraData())

}

func (t BlockTable) addHeader(ctx context.Context, h *types.Header) {
	row := t.GetRowCount()
	t.SetCell(row, 0, cview.NewTableCell(time.Unix(int64(h.Time), 0).String()))
	t.SetCell(row, 1, cview.NewTableCell(h.Number.String()))
	t.SetCell(row, 2, cview.NewTableCell(h.Hash().String()))
	t.SetCell(row, 3, cview.NewTableCell(h.ParentHash.String()))
	t.SetCell(row, 4, cview.NewTableCell(h.Coinbase.String()))
	t.SetCell(row, 5, cview.NewTableCell(fmt.Sprint(h.GasLimit)))
	t.SetCell(row, 6, cview.NewTableCell(fmt.Sprint(h.GasUsed)))
	t.SetCell(row, 7, cview.NewTableCell(h.BaseFee.String()))
	t.SetCell(row, 8, cview.NewTableCell(h.Root.String()))
	t.SetCell(row, 9, cview.NewTableCell(string(h.Extra)))

	return

}

func (t BlockTable) watch(ctx context.Context) error {
	for {
		select {
		case header := <-t.ch:
			t.app.app.QueueUpdateDraw(func() {
				row := newBlockRow(header)
				t.addHeaderRow(row)
				// t.Sort(1, true)
			})
		case <-ctx.Done():
			return nil
		}
	}

}
