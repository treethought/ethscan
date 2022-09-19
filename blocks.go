package main

import (
	"context"
	"fmt"
	"time"

	"code.rocketnine.space/tslocum/cview"
	"github.com/aquilax/truncate"
	"github.com/ethereum/go-ethereum/core/types"
)

const truncSize = 8

type blockRow struct {
	header *types.Header
}

func newBlockRow(h *types.Header) blockRow {
	return blockRow{header: h}
}

func (r blockRow) Time() *cview.TableCell {
	t := time.Unix(int64(r.header.Time), 0)

	human := t.Format("01/02/06 3:04 pm")
	return cview.NewTableCell(human)
}

func (r blockRow) Number() *cview.TableCell {
	return cview.NewTableCell(r.header.Number.String())
}

func (r blockRow) Hash() *cview.TableCell {
	hash := r.header.Hash().String()
	tr := truncate.Truncate(hash, truncSize, "...", truncate.PositionMiddle)
	return cview.NewTableCell(tr)
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
func (r blockRow) ExtraData() *cview.TableCell {
	val := string(r.header.Extra)
	// tr := truncate.Truncate(val, truncSize, "...", truncate.PositionEnd)
	return cview.NewTableCell(val)
}

type BlockTable struct {
	*cview.Table
	app     App
	headers []*types.Header
	fields  []string

	ch chan *types.Header
}

func NewBlockTable(app App) *BlockTable {
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
		"Parent",
		"Miner",
		"GasLimit",
		"GasUsed",
		"BaseFee",
		"StateRoot",
		"ExtraData",
	}

	table.SetBorders(true)
	table.SetFixed(0, 0)
	table.SetSelectable(true, false)

	table.setTableHeader()
	table.ch = app.broker.SubscribeHeaders()
	return table

}

func (t BlockTable) setTableHeader() {
	for col, field := range t.fields {
		t.SetCell(0, col, cview.NewTableCell(field))
	}
}

func (t BlockTable) addHeaderRow(r blockRow) {
	row := t.GetRowCount()
	t.SetCell(row, 0, r.Time())
	t.SetCell(row, 1, r.Number())
	t.SetCell(row, 2, r.Hash())
	t.SetCell(row, 3, r.Parent())
	t.SetCell(row, 4, r.Miner())
	t.SetCell(row, 5, r.GasLimit())
	t.SetCell(row, 6, r.GasUsed())
	t.SetCell(row, 7, r.BaseFee())
	t.SetCell(row, 8, r.StateRoot())
	t.SetCell(row, 9, r.ExtraData())

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
			})
		case <-ctx.Done():
			return nil
		}
	}

}
