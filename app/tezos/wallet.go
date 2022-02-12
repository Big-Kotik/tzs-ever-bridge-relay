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

func (w *Wallet) SendTransaction(ctx context.Context, to tezos.Address) {
	op := codec.NewOp()
	op = op.WithBranch(w.bw.GetLastBlock())

	user, err := w.client.GetContract(ctx, w.key.Address(), w.bw.GetLastBlock())

	if err != nil {
		log.Println(err)
		return
	}

	//TODO: add gas getter
	trs := &codec.Transaction{
		Manager: codec.Manager{
			Source:   w.key.Address(),
			Counter:  tezos.N(user.Counter + 1),
			GasLimit: 100000,
			Fee:      1000,
		},
		Destination: to,
		Amount:      100000,
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
