package main

import (
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"fmt"
	tz "tez-ton-bridge-relay/app/tezos"
)

func main() {
	ctx := context.TODO()

	addr := tezos.MustParseAddress("KT1KDswcMU81pJbLJfWh4JtjNXuo8fd2tbqL")

	c, _ := rpc.NewClient("http://localhost:20000/", nil)

	blockWatcher := tz.NewBlockWatcher(c)
	contractWatcher := tz.NewContractWatcher(c, blockWatcher, addr)

	go blockWatcher.Run(ctx)
	go contractWatcher.Run(ctx)

	for {
		transaction := <-contractWatcher.Transactions

		fmt.Println(tz.ParseFromTransaction(transaction))
	}
}
