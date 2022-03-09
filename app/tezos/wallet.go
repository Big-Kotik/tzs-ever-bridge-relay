package tezos

import (
	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
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

func (w *Wallet) Address(ctx context.Context) (tezos.Address, error) {
	return w.key.Address(), nil
}

func (w *Wallet) Key(ctx context.Context) (tezos.Key, error) {
	return w.key.Public(), nil
}

func (w *Wallet) SignMessage(ctx context.Context, s string) (tezos.Signature, error) {
	return w.key.Sign([]byte(s))
}

func (w *Wallet) SignOperation(ctx context.Context, op *codec.Op) (tezos.Signature, error) {
	return w.key.Sign(op.Digest())
}

func (w *Wallet) SignBlock(ctx context.Context, header *codec.BlockHeader) (tezos.Signature, error) {
	return w.key.Sign(header.Bytes())
}

func NewWallet(client *rpc.Client, bw *BlockWatcher, key tezos.PrivateKey) *Wallet {
	return &Wallet{
		client, bw, key,
	}
}

type TransactionParams struct {
	To     tezos.Address
	Amount tezos.N
	Params *micheline.Parameters
}

func (w *Wallet) SendTransaction(ctx context.Context, transaction TransactionParams) {
	op := codec.NewOp()
	op = op.WithBranch(w.bw.GetLastBlock())

	user, err := w.client.GetContract(ctx, w.key.Address(), w.bw.GetLastBlock())

	if err != nil {
		log.Printf("err: %s\n", err)
		return
	}

	//TODO: add gas getter
	trs := &codec.Transaction{
		Manager: codec.Manager{
			Source:       w.key.Address(),
			Counter:      tezos.N(user.Counter + 1),
			GasLimit:     100000,
			Fee:          2000000,
			StorageLimit: 10000,
		},
		Destination: transaction.To,
		Amount:      transaction.Amount,
	}
	op.WithContents(trs)

	if err := op.Sign(w.key); err != nil {
		log.Printf("err: %s\n", err)
		return
	}

	hash, err := w.client.BroadcastOperation(ctx, op.Bytes())
	if err != nil {
		log.Printf("err: %s\n", string(err.(rpc.HTTPError).Body()))
		return
	}

	log.Printf("Operation hash: %s", hash)
}
