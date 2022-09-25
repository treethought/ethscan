package main

import (
	"context"
	"fmt"

	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TransacionLogs struct {
	*cview.TreeView
	logs []*types.Log
	txn  *types.Transaction
	app  *App
	db   *SignatureDB
	abi  *abi.ABI
}

func NewTransactionLogs(app *App, logs []*types.Log) *TransacionLogs {
	return &TransacionLogs{
		TreeView: cview.NewTreeView(),
		logs:     logs,
		app:      app,
		db:       NewSignatureDB(),
	}
}

func (tl *TransacionLogs) Update() {
	if tl.GetRoot() != nil {
		tl.GetRoot().ClearChildren()
	}
	txn := tl.app.state.txn

	tl.app.app.QueueUpdateDraw(func() {
		rec, err := tl.app.client.TransactionReceipt(context.TODO(), txn.Hash())
		if err != nil {
			tl.app.log.Error("failed to get txn receipt")
			return
		}
		tl.txn = txn
		tl.logs = rec.Logs
		abi, err := GetContractABI(txn.To().String())
		if err != nil {
			tl.app.log.Errorf("failed to get contract abi: %s %s", txn.To().String(), err)
		}
		tl.abi = abi
		tl.render()
	})

}

func (tl *TransacionLogs) buildTopic(topic common.Hash, idx int, checkSig bool) *cview.TreeNode {
	if !checkSig {
		return cview.NewTreeNode(fmt.Sprintf("%d: %s", idx, topic.Hex()))
	}

	// first get method from abi if we have it
	if tl.abi != nil {
		event, err := tl.abi.EventByID(topic)
		if err != nil {
			tl.app.log.Error("failed to get event from abi")
		} else {
			return cview.NewTreeNode(event.Sig)
		}
	}

	// otherwise check for signature

	// get first 4 bytes + 0x
	prefix := topic.Hex()[:10]
	tl.app.log.Debug("checking for signature for hex: ", prefix)
	sig, err := tl.db.GetSignature(prefix)
	if err != nil {
		tl.app.log.Error("failed to get topic signature: ", err)
		return cview.NewTreeNode(fmt.Sprintf("%d: %s", idx, topic.Hex()))

	}
	tl.app.log.Info("got signature for %s", prefix)
	return cview.NewTreeNode(sig.TextSignature)

}

func (tl *TransacionLogs) render() {
	tl.SetTitle("Logs")
	tl.SetRoot(cview.NewTreeNode("Logs"))

	if len(tl.logs) == 0 {
		return
	}

	for _, l := range tl.logs {
		addr := cview.NewTreeNode(fmt.Sprintf("Adress: %s", formatAddress(tl.app.client, l.Address)))

		topics := cview.NewTreeNode("Topics")
		for i, t := range l.Topics {
			var n *cview.TreeNode
			if i == 0 {
				n = tl.buildTopic(t, i, true)
			} else {
				n = tl.buildTopic(t, i, false)
			}
			topics.AddChild(n)
		}
		addr.AddChild(topics)
		tl.GetRoot().AddChild(addr)
	}

}
