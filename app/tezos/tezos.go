package tezos

import (
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"log"
)

type BlockWatcher struct {
	client *rpc.Client
	Hashes chan tezos.BlockHash
}

func NewBlockWatcher(client *rpc.Client) *BlockWatcher {
	return &BlockWatcher{client, make(chan tezos.BlockHash)}
}

func (bw *BlockWatcher) Run(ctx context.Context) {
	mon := rpc.NewBlockHeaderMonitor()
	defer mon.Close()

	if err := bw.client.MonitorBlockHeader(ctx, mon); err != nil {
		log.Fatalln(err)
	}

	for {
		head, err := mon.Recv(ctx)
		if err != nil {
			log.Fatalln(err)
		}

		bw.Hashes <- head.Hash
	}
}

type ContractWatcher struct {
	client       *rpc.Client
	blockWatcher *BlockWatcher
	Transactions chan *rpc.Transaction
	address      tezos.Address
}

func NewContractWatcher(client *rpc.Client, blockWatcher *BlockWatcher, address tezos.Address) *ContractWatcher {
	return &ContractWatcher{client, blockWatcher, make(chan *rpc.Transaction), address}
}

func (cw *ContractWatcher) Run(ctx context.Context) {
	for {
		blockHash := <-cw.blockWatcher.Hashes

		operations, err := cw.client.GetBlockOperations(ctx, blockHash)

		if err != nil {
			log.Fatalln(err)
		}

		for i := range operations {
			for _, operation := range operations[i] {
				for _, content := range operation.Contents {
					if content.Kind() == tezos.OpTypeTransaction {
						transaction := content.(*rpc.Transaction)
						if transaction.Destination.String() == cw.address.String() {
							cw.Transactions <- transaction
						}
					}
				}
			}
		}
	}
}
