package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/aquilax/truncate"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gdamore/tcell/v2"
)

type TransactionTable struct {
	*cview.Table
	app      *App
	txs      []*types.Transaction
	block    *types.Block
	bindings *cbind.Configuration

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
	table.SetSelectable(true, false)

	table.initBindings()

	table.SetSelectedFunc(func(row, _ int) {
		if row == 0 {
			return
		}
		table.app.log.Info("Selected txn - row ", row)

		txn := table.getCurrentRef()
		if txn == nil {
			log.Fatal("reference was not a txn")
		}

		table.app.log.Info("Row reference txn hash: ", txn.Hash().String())
		table.app.ShowTransactonData(txn)
	})

	go table.getTransactions()
	return table
}

func (t *TransactionTable) initBindings() {
	t.bindings = cbind.NewConfiguration()
	t.SetInputCapture(t.bindings.Capture)
	t.bindings.SetRune(tcell.ModNone, 'o', t.handleOpen)

}

func (t *TransactionTable) getCurrentRef() *types.Transaction {
	row, _ := t.GetSelection()
	ref := t.GetCell(row, 0).GetReference()
	txn, ok := ref.(*types.Transaction)
	if !ok {
		t.app.log.Error("failed to get txn ref")
		return nil
	}
	return txn
}

func (t *TransactionTable) handleOpen(ev *tcell.EventKey) *tcell.EventKey {
	cur := t.getCurrentRef()
	url := fmt.Sprintf("https://etherscan.io/tx/%s", cur.Hash().Hex())
	openbrowser(url)
	return nil

}

func (t *TransactionTable) getTransactions() {
	if t.block == nil {
		return
	}
	t.app.log.Info("iterating txns: ", len(t.block.Transactions()))
	for i, txn := range t.block.Transactions() {
		t.app.log.Infof("adding txn num %d: %s", i, txn.Hash().String())
		t.addTxn(context.TODO(), i+1, txn)
	}

}

func (t TransactionTable) setHeader() {
	t.SetCell(0, 0, cview.NewTableCell("Txn Hash"))
	t.SetCell(0, 1, cview.NewTableCell("Block"))
	t.SetCell(0, 2, cview.NewTableCell("Age"))
	t.SetCell(0, 3, cview.NewTableCell("From"))
	t.SetCell(0, 4, cview.NewTableCell("To"))
	t.SetCell(0, 5, cview.NewTableCell("Value (Eth)"))
	t.SetCell(0, 6, cview.NewTableCell("Fee (Eth)"))
	t.SetCell(0, 7, cview.NewTableCell("Status"))
	t.SetFixed(1, 0)
}

func (t TransactionTable) addTxn(ctx context.Context, row int, txn *types.Transaction) {
	if txn == nil {
		return
	}

	msg, err := txn.AsMessage(t.app.signer, t.block.BaseFee())
	if err != nil {
		t.app.log.Error("failed to get txn as message: ", err)
	}

	receipt, err := t.app.client.TransactionReceipt(ctx, txn.Hash())
	if err != nil {
		log.Fatal(err)
	}

	blockTime := time.Unix(int64(t.block.Time()), 0)
	age := time.Since(blockTime).Truncate(time.Second)

	hash := truncate.Truncate(txn.Hash().String(), truncSize, "...", truncate.PositionMiddle)

	statusText := "success"
	if receipt.Status != 1 {
		statusText = "failed"
	}

	t.app.app.QueueUpdateDraw(func() {
		hashRefCell := cview.NewTableCell(hash)
		hashRefCell.SetReference(txn)
		t.SetCell(row, 0, hashRefCell)
		t.SetCell(row, 1, cview.NewTableCell(receipt.BlockNumber.String()))
		t.SetCell(row, 2, cview.NewTableCell(age.String()))
		t.SetCell(row, 3, cview.NewTableCell(msg.From().Hex()))
		t.SetCell(row, 4, cview.NewTableCell(txn.To().Hex()))
		t.SetCell(row, 5, cview.NewTableCell(weiToEther(txn.Value()).String()))
		fee := getFee(receipt, txn, t.block.BaseFee())
		t.SetCell(row, 6, cview.NewTableCell(weiToEther(fee).String()))
		t.SetCell(row, 7, cview.NewTableCell(statusText))
	})
	// go t.resolveAddresses(row)
}

func (t TransactionTable) resolveAddresses(row int) {
	t.app.log.Info("resolving row: ", row)
	from := t.GetCell(row, 3)
	to := t.GetCell(row, 4)
	t.app.app.QueueUpdateDraw(func() {
		from.SetText(formatAddress(t.app.client, common.HexToAddress(from.GetText())))
		to.SetText(formatAddress(t.app.client, common.HexToAddress(to.GetText())))
	})
}
