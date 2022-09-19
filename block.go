package main

import (
	"fmt"

	"code.rocketnine.space/tslocum/cview"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockData struct {
	*cview.Flex
	client *ethclient.Client
	block  *types.Block
}

func NewBlockData(client *ethclient.Client, block *types.Block) *BlockData {
	bd := &BlockData{client: client, block: block, Flex: cview.NewFlex()}
	bd.render()
	return bd
}

func (d BlockData) blockHeaders() *cview.List {
	l := cview.NewList()
	l.SetTitle("Headers")

	number := d.block.Number().String()
	hash := d.block.Hash().String()
	time := formatUnixTime(d.block.Time())
	parent := d.block.ParentHash().String()
	coinbase := d.block.Coinbase().String()
	gasLimit := d.block.GasLimit()
	gasUsed := d.block.GasUsed()
	baseFee := d.block.BaseFee()
	root := d.block.Root().String()
	extraData := string(d.block.Extra())

	l.AddItem(cview.NewListItem(fmt.Sprintf("Number: %s", number)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Hash: %s", hash)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Time: %s", time)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Parent: %s", parent)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Coinbase (Proposer): %s", coinbase)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("GasLimit: %d", gasLimit)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("GasUsed: %d", gasUsed)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("BaseFee: %s", baseFee)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("Root: %s", root)))
	l.AddItem(cview.NewListItem(fmt.Sprintf("ExtraData: %s", extraData)))
	return l

}

func (d BlockData) render() {
	// headers

	d.AddItem(d.blockHeaders(), 0, 1, false)
	d.AddItem(nil, 0, 1, false)

	txns := cview.NewList()
	txns.SetTitle("Transactions")
	d.AddItem(txns, 0, 3, true)

	d.SetTitle(fmt.Sprintf("Block #%s", d.block.Number().String()))
}
