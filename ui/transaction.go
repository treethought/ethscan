package ui

import (
	"context"
	"fmt"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/treethought/ethscan/util"
)

type TransactionData struct {
	*cview.Grid
	app      *App
	bindings *cbind.Configuration
	txn      *types.Transaction
	block    *types.Block
}

func NewTransactionData(app *App, txn *types.Transaction) *TransactionData {
	d := &TransactionData{
		Grid: cview.NewGrid(),
		app:  app,
		txn:  txn,
	}
	return d

}

func (d *TransactionData) Update() {
	txn := d.app.State.txn
	d.block = d.app.State.block
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

	msg, err := d.txn.AsMessage(d.app.signer, d.block.BaseFee())
	if err != nil {
		d.app.log.Error("failed to get txn as message: ", err)
	}

	meta := cview.NewList()
	meta.SetTitle("meta")
	meta.SetBorder(true)

	hash := cview.NewListItem("Hash")
	hash.SetSecondaryText(d.txn.Hash().Hex())
	meta.AddItem(hash)

	status := cview.NewListItem("Status")
	statusText := "Success"
	if rec.Status != 1 {
		statusText = "[red]Failed[red]"
	}
	status.SetSecondaryText(statusText)
	meta.AddItem(status)

	block := cview.NewListItem("Block")
	block.SetSecondaryText(rec.BlockNumber.String())
	meta.AddItem(block)

	info := cview.NewList()
	info.SetTitle("info")
	info.SetBorder(true)

	from := cview.NewListItem("From")
	from.SetSecondaryText(util.FormatAddress(d.app.client, msg.From()))
	info.AddItem(from)

	var to *cview.ListItem
	// contact
	if len(d.txn.Data()) > 0 {
		to = cview.NewListItem("Interacted With (To)")
	} else {
		to = cview.NewListItem("To")
	}
	if msg.To() != nil {
		to.SetSecondaryText(util.FormatAddress(d.app.client, *msg.To()))
	} else {
		to.SetSecondaryText("none")
	}
	info.AddItem(to)

	value := cview.NewListItem("Value")
	val := util.WeiToEther(d.txn.Value())
	valText := fmt.Sprintf("%s ETH", val.String())
	value.SetSecondaryText(valText)
	info.AddItem(value)

	fee := cview.NewListItem("Transaction Fee")
	feeWei := util.GetFee(rec, d.txn, d.block.BaseFee())
	feeStr := fmt.Sprintf("%s ETH", util.WeiToEther(feeWei).String())

	fee.SetSecondaryText(feeStr)
	info.AddItem(fee)

	gas := cview.NewList()
	gas.SetTitle("Gas")
	gas.SetBorder(true)

	gasPrice := cview.NewListItem("Gas Price (Gwei)")
	gasPrice.SetSecondaryText(util.WeiToGwei(msg.GasPrice()).String())
	gas.AddItem(gasPrice)

	gasLimit := cview.NewListItem("Gas Limit")
	gasLimit.SetSecondaryText(fmt.Sprint(msg.Gas()))
	gas.AddItem(gasLimit)

	pctUsed := float64(rec.GasUsed) / float64(msg.Gas()) * float64(100)
	gasUsed := cview.NewListItem("Gas Used")
	gasUsed.SetSecondaryText(fmt.Sprintf("%d (%.2f)%%", rec.GasUsed, pctUsed))
	gas.AddItem(gasUsed)

	gasFees := cview.NewListItem("Gas Fees (Gwei)")
	gasFeeTxt := fmt.Sprintf("Base: %.2f | Max: %.2f | Max Priority: %.2f",
		util.WeiToGwei(d.block.BaseFee()),
		util.WeiToGwei(msg.GasFeeCap()),
		util.WeiToGwei(msg.GasTipCap()))

	gasFees.SetSecondaryText(gasFeeTxt)
	gas.AddItem(gasFees)

	other := cview.NewList()
	other.SetTitle("other")
	other.SetBorder(true)

	txnType := cview.NewListItem("Txn Type")
	var txnTypeTxt string
	switch d.txn.Type() {
	case 0:
		txnTypeTxt = "0 (Legacy)"
	case 1:
		txnTypeTxt = "1 (Contract Deployment)"
	case 2:
		txnTypeTxt = "2 (EIP-1559)"
	}
	txnType.SetSecondaryText(txnTypeTxt)

	nonce := cview.NewListItem("Nonce")
	nonce.SetSecondaryText(fmt.Sprint(d.txn.Nonce()))

	pos := cview.NewListItem("Position")
	pos.SetSecondaryText(fmt.Sprint(rec.TransactionIndex))

	other.AddItem(txnType)
	other.AddItem(nonce)
	other.AddItem(pos)

	data := cview.NewTextView()
	data.SetTitle("Input Data")
	data.SetBorder(true)

	data.SetText(string(rec.PostState))

	d.SetRows(0, 0)
	d.SetColumns(0, 0, 0, 0)

	d.AddItem(meta, 0, 0, 1, 2, 0, 0, false)
	d.AddItem(info, 0, 2, 1, 2, 0, 0, false)
	d.AddItem(gas, 1, 0, 1, 2, 0, 0, false)
	d.AddItem(other, 1, 2, 1, 2, 0, 0, false)

}
