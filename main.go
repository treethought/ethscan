package main

import (
	"log"
	"os"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial(os.Getenv("ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}

	app := NewApp(client)
	app.Start()

}
