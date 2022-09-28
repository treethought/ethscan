package util

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

func GetFee(rec *types.Receipt, txn *types.Transaction, baseFee *big.Int) *big.Int {
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
