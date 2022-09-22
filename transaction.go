package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransactionTable struct {
	*cview.Table
	app   *App
	txs   []*types.Transaction
	block *types.Block

	ch chan *types.Transaction
}

func NewTransactionTable(app *App, block *types.Block) *TransactionTable {
	table := &TransactionTable{
		Table: cview.NewTable(),
		txs:   nil,
		ch:    make(chan *types.Transaction),
		app:   app,
		block: block,
	}
	table.SetBorder(true)

	table.setHeader()
	go table.getTransactions()
	// table.ch = app.broker.SubscribeTransactions()
	return table
}

func (t *TransactionTable) getTransactions() {
	if t.block == nil {
		return
	}
	t.app.log.Info("iterating txns: ", len(t.block.Transactions()))
	for i, txn := range t.block.Transactions() {
		t.app.app.QueueUpdateDraw(func() {
			t.app.log.Infof("adding txn num %d: %s", i, txn.Hash().String())
			t.addTxn(context.TODO(), txn)
		})
	}

}

func (t TransactionTable) setHeader() {
	t.SetCell(0, 0, cview.NewTableCell("Txn Hash"))
	t.SetCell(0, 1, cview.NewTableCell("Method"))
	t.SetCell(0, 2, cview.NewTableCell("Block"))
	t.SetCell(0, 3, cview.NewTableCell("Age"))
	t.SetCell(0, 4, cview.NewTableCell("From"))
	t.SetCell(0, 5, cview.NewTableCell("To"))
	t.SetCell(0, 6, cview.NewTableCell("Value"))
	t.SetCell(0, 7, cview.NewTableCell("Fee"))
	t.SetCell(0, 8, cview.NewTableCell("Status"))
}

func (t TransactionTable) addTxn(ctx context.Context, txn *types.Transaction) {

	receipt, err := t.app.client.TransactionReceipt(ctx, txn.Hash())
	if err != nil {
		log.Fatal(err)
	}

	block, err := t.app.client.BlockByHash(ctx, receipt.BlockHash)
	if err != nil {
		log.Fatal(err)
	}

	blockTime := time.Unix(int64(block.Time()), 0)
	age := time.Since(blockTime)

	row := t.GetRowCount()
	t.SetCell(row, 0, cview.NewTableCell(txn.Hash().String()))
	t.SetCell(row, 1, cview.NewTableCell("?"))
	t.SetCell(row, 2, cview.NewTableCell(receipt.BlockNumber.String()))
	t.SetCell(row, 3, cview.NewTableCell(age.String()))
	t.SetCell(row, 4, cview.NewTableCell("?"))
	t.SetCell(row, 5, cview.NewTableCell("To"))
	t.SetCell(row, 6, cview.NewTableCell(txn.Value().String()))
	t.SetCell(row, 7, cview.NewTableCell("?"))
	t.SetCell(row, 8, cview.NewTableCell(fmt.Sprint(receipt.Status)))

}

// func (t TransactionTable) watch(ctx context.Context) error {
// 	for {
// 		select {
// 		case tx := <-t.ch:
// 			t.app.app.QueueUpdateDraw(func() {

// 				t.txs = append(t.txs, tx)
// 				t.addTxn(ctx, tx)
// 			})
// 		case <-ctx.Done():
// 			return nil
// 		}
// 	}

// }
