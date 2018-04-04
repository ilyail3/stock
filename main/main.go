package main

import (
	"fmt"
	"github.com/ilyail3/stock"
	"os"
)

func main(){
	config, err := stock.ReadConfig(os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Printf("api key:%s\n", config.GetApiKey())
	reader := stock.AlphaAdvantageReader(config.GetApiKey())

	r, err := reader.GetPrice("TSLA")

	if err != nil {
		panic(err)
	}

	fmt.Println(r)

	r, err = reader.GetPrice("DPS")

	if err != nil {
		panic(err)
	}

	fmt.Println(r)
}