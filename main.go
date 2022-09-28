package main

import (
	"log"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/treethought/ethscan/ui"
)

func main() {
	client, err := ethclient.Dial(os.Getenv("ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}

	app := ui.NewApp(client)
	app.Start()

}
