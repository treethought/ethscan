package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"code.rocketnine.space/tslocum/cbind"
	"code.rocketnine.space/tslocum/cview"
	"github.com/aquilax/truncate"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gdamore/tcell/v2"
)

type abiMsg struct {
	txn *types.Transaction
	abi *abi.ABI
	row int
}

type TransactionTable struct {
	*cview.Table
	app      *App
	txs      []*types.Transaction
	block    *types.Block
	bindings *cbind.Configuration

	abiChan chan abiMsg
}

func NewTransactionTable(app *App, block *types.Block) *TransactionTable {
	table := &TransactionTable{
		Table:   cview.NewTable(),
		txs:     nil,
		abiChan: make(chan abiMsg),
		app:     app,
		block:   block,
	}
	table.SetBorder(true)

	table.setHeader()
	table.SetSelectable(true, false)

	table.initBindings()

	go table.handleAbis(context.TODO())
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
		table.app.state.SetBlock(block)
		table.app.state.SetTxn(txn)
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
		t.addTxn(context.TODO(), i+1, txn)
	}

}

func (t TransactionTable) setHeader() {
	t.SetCell(0, 0, cview.NewTableCell("Txn Hash"))
	t.SetCell(0, 1, cview.NewTableCell("Method"))
	t.SetCell(0, 2, cview.NewTableCell("Age"))
	t.SetCell(0, 3, cview.NewTableCell("From"))
	t.SetCell(0, 4, cview.NewTableCell("To"))
	t.SetCell(0, 5, cview.NewTableCell("Value (Eth)"))
	t.SetCell(0, 6, cview.NewTableCell("Fee (Eth)"))
	t.SetCell(0, 7, cview.NewTableCell("Logs"))
	t.SetCell(0, 8, cview.NewTableCell("Status"))
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

	go func() {
		from := formatAddress(t.app.client, msg.From())
		var to string
		if txn.To() == nil {
			to = ""
		} else {
			to = formatAddress(t.app.client, *txn.To())
		}
		t.app.app.QueueUpdateDraw(func() {
			fromCell := t.GetCell(row, 3)
			fromCell.SetText(from)
			toCell := t.GetCell(row, 4)
			toCell.SetText(to)
		})
	}()

	receipt, err := t.app.client.TransactionReceipt(ctx, txn.Hash())
	if err != nil {
		log.Fatal(err)
	}

	var method string
	var toField string

	// contract deployment
	if txn.To() == nil {
		method = "[red]ContractDeployment[red]"
		toField = ""
	} else if len(txn.Data()) >= 4 {
		toField = txn.To().Hex()
		// contract execution
		method = common.Bytes2Hex(txn.Data()[:4])
		go func() {
			abi, err := GetContractABI(toField)
			if err != nil {
				t.app.log.Errorf("failed to get abi: ", err)
			}
			t.abiChan <- abiMsg{abi: abi, txn: txn, row: row}
		}()
	} else {
		toField = txn.To().Hex()
		// basic eth transfer
		method = fmt.Sprintf("[yellow]Transfer[yellow]")
	}

	blockTime := time.Unix(int64(t.block.Time()), 0)
	age := time.Since(blockTime).Truncate(time.Second)

	hash := truncate.Truncate(txn.Hash().String(), truncSize, "...", truncate.PositionMiddle)

	statusText := "success"
	if receipt.Status != 1 {
		statusText = "failed"
	}

	// to may be nil if is contract deployment

	t.app.app.QueueUpdateDraw(func() {
		hashRefCell := cview.NewTableCell(hash)
		hashRefCell.SetReference(txn)
		t.SetCell(row, 0, hashRefCell)

		methodCell := cview.NewTableCell(method)
		t.SetCell(row, 1, methodCell)
		t.SetCell(row, 2, cview.NewTableCell(age.String()))
		t.SetCell(row, 3, cview.NewTableCell(msg.From().Hex()))
		t.SetCell(row, 4, cview.NewTableCell(toField))
		t.SetCell(row, 5, cview.NewTableCell(weiToEther(txn.Value()).String()))
		fee := getFee(receipt, txn, t.block.BaseFee())
		t.SetCell(row, 6, cview.NewTableCell(weiToEther(fee).String()))
		t.SetCell(row, 7, cview.NewTableCell(fmt.Sprint(len(receipt.Logs))))
		t.SetCell(row, 8, cview.NewTableCell(statusText))
	})
	// go t.resolveAddresses(row)
}

func (t *TransactionTable) resolveAddresses(row int) {
	t.app.log.Info("resolving row: ", row)
	from := t.GetCell(row, 3)
	to := t.GetCell(row, 4)
	t.app.app.QueueUpdateDraw(func() {
		from.SetText(formatAddress(t.app.client, common.HexToAddress(from.GetText())))
		to.SetText(formatAddress(t.app.client, common.HexToAddress(to.GetText())))
	})
}
func (t *TransactionTable) setMethod(m abiMsg) {
	if m.abi == nil || len(m.txn.Data()) == 0 {
		return
	}
	method, _, err := DecodeTransactionInputData(m.abi, m.txn.Data())
	if err != nil {
		t.app.log.Error("failed to decode txn input data: ", err)
	}

	cell := t.GetCell(m.row, 1)
	cell.SetText(fmt.Sprintf("[blue]%s[blue]", method))

}

func (t *TransactionTable) handleAbis(ctx context.Context) {
	for {
		select {
		case m := <-t.abiChan:
			t.app.log.Infof("received abi for addr: %s", m.txn.To().String())
			t.setMethod(m)
		case <-ctx.Done():
			return

		}
	}
}
