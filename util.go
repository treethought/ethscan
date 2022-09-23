package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os/exec"
	"runtime"
	"time"

	"github.com/aquilax/truncate"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	ens "github.com/wealdtech/go-ens/v3"
)

func formatAddress(client *ethclient.Client, addr common.Address) string {
	name, err := ens.ReverseResolve(client, addr)
	if err == nil {
		return name
	}
	return truncate.Truncate(addr.Hex(), truncSize, "...", truncate.PositionMiddle)
}

func formatTime(t time.Time) string {
	return t.Format("01/02/06 15:04")
}

func formatUnixTime(t uint64) string {
	tm := time.Unix(int64(t), 0)
	return formatTime(tm)
}

func weiToEther(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.Ether))
}

func getSigner(ctx context.Context, client *ethclient.Client) types.Signer {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	signer := types.NewLondonSigner(chainID)
	return signer
}

func getFee(rec *types.Receipt, txn *types.Transaction, baseFee *big.Int) *big.Int {
	switch txn.Type() {
	case 2:
		return GetEIP1559Fee(rec, txn, baseFee)
	case 0:
		return GetLegacyFee(txn, rec)
	default:
		return big.NewInt(0)

	}

}

func GetEIP1559Fee(rec *types.Receipt, txn *types.Transaction, baseFee *big.Int) *big.Int {
	feeTip := new(big.Int).Add(baseFee, txn.GasTipCap())
	return new(big.Int).Mul(feeTip, new(big.Int).SetUint64(rec.GasUsed))

}

func GetLegacyFee(txn *types.Transaction, rec *types.Receipt) *big.Int {
	return new(big.Int).Mul(txn.GasPrice(), new(big.Int).SetUint64(rec.GasUsed))

}

func openbrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}
