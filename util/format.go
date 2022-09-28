package util

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"

	ens "github.com/wealdtech/go-ens/v3"
)

func HexStripZeros(hex string) string {
	trimmed := strings.TrimPrefix(hex, "0x")
	trimmed = strings.TrimLeft(trimmed, "0")
	if len(trimmed) < 2 {
		trimmed = fmt.Sprintf("0%s", trimmed)
	}
	return fmt.Sprintf("0x%s", trimmed)
}

func FormatAddress(client *ethclient.Client, addr common.Address) string {
	known, ok := commonAddresses[strings.ToLower(addr.Hex())]
	if ok {
		return fmt.Sprintf("[orange]%s[orange]", known)
	}
	name, err := ens.ReverseResolve(client, addr)
	if err == nil {
		return fmt.Sprintf("[blue]%s[blue]", name)
	}
	return addr.Hex()
}

func FormatTime(t time.Time) string {
	return t.Format("01/02/06 15:04")
}

func FormatUnixTime(t uint64) string {
	tm := time.Unix(int64(t), 0)
	return FormatTime(tm)
}

func WeiToGwei(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.GWei))
}

func WeiToEther(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.Ether))
}
