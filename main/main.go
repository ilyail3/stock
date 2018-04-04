package main

import (
	"fmt"
	"github.com/ilyail3/stock"
)

func main(){
	config, err := stock.ReadConfig("config.json")
	fmt.Printf("api key:%s\n", config.GetApiKey())

	if err != nil {
		panic(err)
	}

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