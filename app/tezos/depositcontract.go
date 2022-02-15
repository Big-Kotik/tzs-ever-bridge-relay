package tezos

import (
	"blockwatch.cc/tzgo/rpc"
)

type WrapTransaction struct {
	amount uint64
	to     string
}

func ParseFromTransaction(transaction *rpc.Transaction) *WrapTransaction {
	if transaction.Parameters == nil {
		return nil
	}
	var wt *WrapTransaction
	switch transaction.Parameters.Entrypoint {
	case "default":
		wt.to = transaction.Parameters.Value.Args[0].Args[0].String
	case "deposit":
		wt.to = transaction.Parameters.Value.String
	default:
		return wt
	}
	wt.amount = uint64(transaction.Amount)
	return wt
}
