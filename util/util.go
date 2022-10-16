package util

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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
	"0xDAFEA492D9c6733ae3d56b7Ed1ADB60692c98Bc5": "Flashbots: Builder",
}

func GetSigner(ctx context.Context, client *ethclient.Client) types.Signer {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	signer := types.NewLondonSigner(chainID)
	return signer
}

func Openbrowser(url string) error {
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
