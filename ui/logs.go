package ui

import (
	"context"
	"fmt"

	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/treethought/ethscan/util"
)

type TransacionLogs struct {
	*cview.TreeView
	logs []*types.Log
	txn  *types.Transaction
	app  *App
	db   *util.SignatureDB
	abi  *abi.ABI
}

func NewTransactionLogs(app *App, logs []*types.Log) *TransacionLogs {
	return &TransacionLogs{
		TreeView: cview.NewTreeView(),
		logs:     logs,
		app:      app,
		db:       util.NewSignatureDB(),
	}
}

func (tl *TransacionLogs) Update() {
	if tl.GetRoot() != nil {
		tl.GetRoot().ClearChildren()
	}
	tl.SetRoot(cview.NewTreeNode("Loading logs..."))
	txn := tl.app.State.txn

	rec, err := tl.app.client.TransactionReceipt(context.TODO(), txn.Hash())
	if err != nil {
		tl.app.log.Error("failed to get txn receipt")
		return
	}
	tl.txn = txn
	tl.logs = rec.Logs
	abi, err := util.GetContractABI(txn.To().String(), tl.app.config.EtherscanKey)
	if err != nil {
		tl.app.log.Errorf("failed to get contract abi: %s %s", txn.To().String(), err)
	}
	tl.abi = abi
	tl.app.app.QueueUpdateDraw(func() {
		tl.render()
	})

}

func (tl *TransacionLogs) decodeLogData(log *types.Log) *cview.TreeNode {
	if tl.abi == nil {
		return nil
	}

	event, err := tl.abi.EventByID(log.Topics[0])
	if err != nil {
		tl.app.log.Error("failed to get log event")
		return nil
	}
	output := make(map[string]interface{})
	err = tl.abi.UnpackIntoMap(output, event.Name, log.Data)
	if err != nil {
		tl.app.log.Error("failed to get log event output")
	}
	n := cview.NewTreeNode("Data")
	for k, v := range output {
		txt := fmt.Sprintf("%s: %v", k, v)
		n.AddChild(cview.NewTreeNode(txt))
	}
	return n

}

func (tl *TransacionLogs) buildTopic(topic common.Hash, idx int, checkSig bool) *cview.TreeNode {
	if !checkSig {
		trimmed := util.HexStripZeros(topic.Hex())
		return cview.NewTreeNode(fmt.Sprintf("%d: %s", idx, trimmed))
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
	tl.app.log.Debug("got signature for %s", prefix)
	return cview.NewTreeNode(sig.TextSignature)

}

func (tl *TransacionLogs) render() {
	tl.SetTitle("Logs")
	tl.SetBorder(true)
	tl.SetRoot(cview.NewTreeNode("."))

	if len(tl.logs) == 0 {
		return
	}

	for _, l := range tl.logs {
		addr := cview.NewTreeNode(fmt.Sprintf("Address: %s", util.FormatAddress(tl.app.client, l.Address)))

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

		data := tl.decodeLogData(l)
		if data != nil {
			addr.AddChild(data)
		}

		tl.GetRoot().AddChild(addr)
	}

}
