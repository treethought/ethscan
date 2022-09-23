package main

import (
	"context"
	"math/big"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransactionData struct {
	cview.List
	app      *App
	bindings *cbind.Configuration
	txn      *types.Transaction
}

func NewTransactionData(app *App, txn *types.Transaction) *TransactionData {
	d := &TransactionData{
		List: *cview.NewList(),
		app:  app,
		txn:  txn,
	}
	return d

}

func (d *TransactionData) SetTransaction(txn *types.Transaction) {
	d.app.app.QueueUpdateDraw(func() {
		if d.txn != nil && d.txn.Hash() == txn.Hash() {
			return
		}
		d.Clear()
		d.txn = txn
		d.render()
	})

}

func (d *TransactionData) render() {
	if d.txn == nil {
		d.Clear()
		return
	}
	rec, err := d.app.client.TransactionReceipt(context.TODO(), d.txn.Hash())
	if err != nil {
		d.app.log.Error("failed to get txn receipt: ", err)
	}

	// TODO: get base fee
	msg, err := d.txn.AsMessage(d.app.signer, big.NewInt(0))
	if err != nil {
		d.app.log.Error("failed to get txn as message: ", err)
	}

	hash := cview.NewListItem("Hash")
	hash.SetSecondaryText(d.txn.Hash().Hex())
	d.AddItem(hash)

	status := cview.NewListItem("Status")
	statusText := "Success"
	if rec.Status != 1 {
		statusText = "Failed"
	}
	status.SetSecondaryText(statusText)
	d.AddItem(status)

	block := cview.NewListItem("Block")
	block.SetSecondaryText(rec.BlockNumber.String())
	d.AddItem(block)

	from := cview.NewListItem("From")
	from.SetSecondaryText(formatAddress(d.app.client, msg.From()))
	d.AddItem(from)

	to := cview.NewListItem("To")
	to.SetSecondaryText(formatAddress(d.app.client, *msg.To()))
	d.AddItem(to)

}
