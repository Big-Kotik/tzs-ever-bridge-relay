package tezos

import (
	"blockwatch.cc/tzgo/rpc"
	"fmt"
	"log"
)

type WrapTransaction struct {
	amount uint64
	to     string
}

func (t *WrapTransaction) String() string {
	return fmt.Sprintf("%d %s", t.amount, t.to)
}

func ParseFromTransaction(transaction *rpc.Transaction) *WrapTransaction {
	if transaction.Parameters == nil {
		return nil
	}
	wt := &WrapTransaction{}
	switch transaction.Parameters.Entrypoint {
	case "default":
		wt.to = transaction.Parameters.Value.Args[0].Args[0].String
	case "deposit":
		log.Println(transaction.Parameters.Value.String)
		wt.to = transaction.Parameters.Value.String
	default:
		return wt
	}
	wt.amount = uint64(transaction.Amount)
	return wt
}
