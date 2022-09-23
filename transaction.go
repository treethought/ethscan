package main

import (
	"context"
	"fmt"
	"math/big"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransacionLogs struct {
	*cview.TreeView
	logs []*types.Log
	app  *App
}

func NewTransactionLogs(app *App, logs []*types.Log) *TransacionLogs {
	return &TransacionLogs{
		TreeView: cview.NewTreeView(),
		logs:     logs,
		app:      app,
	}
}

func (tl *TransacionLogs) Update() {
	txn := tl.app.state.txn

	tl.app.app.QueueUpdateDraw(func() {
		rec, err := tl.app.client.TransactionReceipt(context.TODO(), txn.Hash())
		if err != nil {
			tl.app.log.Error("failed to get txn receipt")
			return
		}
		tl.logs = rec.Logs
		tl.render()
	})

}

func (tl *TransacionLogs) render() {
	tl.SetTitle("Logs")
	if tl.GetRoot() != nil {
		tl.GetRoot().ClearChildren()
	}
	tl.SetRoot(cview.NewTreeNode("Logs"))

	if len(tl.logs) == 0 {
		return
	}
	for _, l := range tl.logs {
		addr := cview.NewTreeNode(fmt.Sprintf("Adress: %s", formatAddress(tl.app.client, l.Address)))

		abi := cview.NewTreeNode(fmt.Sprintf("ABI: %s", string(l.Data)))

		topics := cview.NewTreeNode("Topics")
		for i, t := range l.Topics {
			n := cview.NewTreeNode(fmt.Sprintf("%d: %s", i, t.Hex()))
			topics.AddChild(n)
		}
		addr.AddChild(abi)
		addr.AddChild(topics)
		tl.GetRoot().AddChild(addr)
	}

}

type TransactionData struct {
	*cview.List
	app      *App
	bindings *cbind.Configuration
	txn      *types.Transaction
}

func NewTransactionData(app *App, txn *types.Transaction) *TransactionData {
	d := &TransactionData{
		List: cview.NewList(),
		app:  app,
		txn:  txn,
	}
	return d

}

func (d *TransactionData) Update() {
	txn := d.app.state.txn
	d.SetTransaction(txn)
}

func (d *TransactionData) SetTransaction(txn *types.Transaction) {
	if d.txn != nil && d.txn.Hash() == txn.Hash() {
		return
	}
	d.app.app.QueueUpdateDraw(func() {
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
