package tezos

import (
	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"log"
)

type Wallet struct {
	client *rpc.Client
	bw     *BlockWatcher
	key    tezos.PrivateKey
}

func NewWallet(client *rpc.Client, bw *BlockWatcher, key tezos.PrivateKey) *Wallet {
	return &Wallet{
		client, bw, key,
	}
}

func (w *Wallet) SendUnwrapTransaction(ctx context.Context, transaction UnwrapTransaction) {
	//TODO: make send transaction func
	op := codec.NewOp()
	op = op.WithBranch(w.bw.GetLastBlock())

	user, err := w.client.GetContract(ctx, w.key.Address(), w.bw.GetLastBlock())

	if err != nil {
		log.Println(err)
		return
	}

	to := tezos.MustParseAddress(transaction.Address)
	//TODO: add gas getter
	trs := &codec.Transaction{
		Manager: codec.Manager{
			Source:   w.key.Address(),
			Counter:  tezos.N(user.Counter + 1),
			GasLimit: 100000,
			Fee:      10000,
		},
		Destination: to,
		Amount:      tezos.N(transaction.Amount),
	}
	op.WithContents(trs)

	if err := op.Sign(w.key); err != nil {
		log.Println(err)
		return
	}

	hash, err := w.client.BroadcastOperation(ctx, op.Bytes())
	if err != nil {
		log.Println(string(err.(rpc.HTTPError).Body()))
		return
	}

	log.Printf("Operation hash: %s", hash)
}
