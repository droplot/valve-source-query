package main

import (
	"fmt"
	"github.com/icraftltd/valve-source-query/source"
)

func main() {
	client, err := source.NewClient("103.205.253.202:36244")

	if err != nil {
		panic(err)
	}

	resp, err := client.Players()
	if err != nil {
		panic(err)
	}

	for k, v := range resp.Items {
		fmt.Println(k, "=", v)
	}
}
