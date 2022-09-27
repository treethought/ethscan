package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	ens "github.com/wealdtech/go-ens/v3"
)

var commonAddresses = map[string]string{
	"0xdac17f958d2ee523a2206206994597c13d831ec7": "Tether USD (USDT)",
	"0x00000000219ab540356cbb839cbe05303d7705fa": "Eth2 Deposit Contract",
	"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2": "Wrapped Ether (WETH)",
	"0xf977814e90da44bfa03b6295a0616a897441acec": "Binance 8",
	"0x2faf487a4414fe77e2327f0bf4ae2a264a776ad2": "FTX Exchange",
	"0x25eaff5b179f209cf186b1cdcbfa463a69df4c45": "Blockfolio",
	"0xdb044b8298e04d442fdbe5ce01b8cc8f77130e33": "Bitkub Hot Wallet 1",
	"0x7a250d5630b4cf539739df2c5dacb4c659f2488d": "Uniswap V2: Router 2",
	"0x68b3465833fb72a70ecdf485e0e4c7bd8665fc45": "Uniswap V3: router 2",
	"0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2": "Maker Token",
	"0x881d40237659c251811cec9c364ef91dc08d300c": "Metamask Swap Router",
	"0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9": "Aave: AAVE Token",
	"0x500a746c9a44f68fe6aa86a92e7b3af4f322ae66": "Voyager 1",
	"0x00000000006c3852cbef3e08e8df289169ede581": "Seaport 1.1",
	"0x1111111254fb6c44bac0bed2854e76f90643097d": "1inch v4: Router",
	"0x46340b20830761efd32832a74d7169b29feb9758": "Crypto.com 2",
	"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": "Centre: USD Coin",
	"0xb5d85cbf7cb3ee0d56b3bb207d5fc4b82f43f511": "Coinbase 5",
	"0x3cd751e6b0078be393132286c442345e5dc49699": "Coinbase 4",
	"0xfd54078badd5653571726c3370afb127351a6f26": "Huobi 30",
	"0x4103c267fba03a1df4fe84bc28092d629fa3f422": "Umbria: Narni Bridge",
	"0xc098b2a3aa256d2140208c3de6543aaef5cd3a94": "FTX Exchange 2",
}

func hexStripZeros(hex string) string {
	trimmed := strings.TrimPrefix(hex, "0x")
	trimmed = strings.TrimLeft(trimmed, "0")
	if len(trimmed) < 2 {
		trimmed = fmt.Sprintf("0%s", trimmed)
	}
	return fmt.Sprintf("0x%s", trimmed)
}

func formatAddress(client *ethclient.Client, addr common.Address) string {
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

func formatTime(t time.Time) string {
	return t.Format("01/02/06 15:04")
}

func formatUnixTime(t uint64) string {
	tm := time.Unix(int64(t), 0)
	return formatTime(tm)
}

func weiToGwei(wei *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	return f.Quo(fWei.SetInt(wei), big.NewFloat(params.GWei))
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
