package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

func formatTime(t time.Time) string {
	return t.Format("01/02/06 3:04 pm")
}

func formatUnixTime(t uint64) string {
	tm := time.Unix(int64(t), 0)
	return formatTime(tm)
}

func ToDecimal(ivalue interface{}, decimals int) decimal.Decimal {
	value := new(big.Int)
	switch v := ivalue.(type) {
	case string:
		value.SetString(v, 10)
	case *big.Int:
		value = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(value.String())
	result := num.Div(mul)

	return result
}

func toEth(w *big.Int) *big.Int {
	return new(big.Int).Div(w, big.NewInt(1000000000000000000))
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
