package tezos

import (
	"blockwatch.cc/tzgo/rpc"
	"fmt"
	"math/big"
)

type WrapTransaction struct {
	amount uint64
	to     string
	To     *big.Int
}

func (t *WrapTransaction) ToHashFormat() string {
	return fmt.Sprintf("0:%s-%d", t.to, t.amount)
}

func ParseFromTransaction(transaction *rpc.Transaction) *WrapTransaction {
	if transaction.Parameters == nil {
		return nil
	}
	wt := &WrapTransaction{}
	switch transaction.Parameters.Entrypoint {
	case "default":
		wt.to = transaction.Parameters.Value.Args[0].Args[0].String
		wt.To = new(big.Int)
		wt.To.SetString(wt.to, 16)
	case "deposit":
		wt.to = transaction.Parameters.Value.String
		wt.To = new(big.Int)
		wt.To.SetString(transaction.Parameters.Value.String, 16)
	default:
		return wt
	}
	wt.amount = uint64(transaction.Amount)
	return wt
}
