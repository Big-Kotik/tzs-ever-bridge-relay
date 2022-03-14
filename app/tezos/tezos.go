package tezos

import (
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"log"
)

type BlockWatcher struct {
	client    *rpc.Client
	lastBlock tezos.BlockHash
}

func NewBlockWatcher(client *rpc.Client, headBlock tezos.BlockHash) *BlockWatcher {
	return &BlockWatcher{client, headBlock}
}

func (bw *BlockWatcher) Run(ctx context.Context, blockHashes chan<- tezos.BlockHash) {
	mon := rpc.NewBlockHeaderMonitor()
	defer mon.Close()
	defer close(blockHashes)

	if err := bw.client.MonitorBlockHeader(ctx, mon); err != nil {
		log.Fatalln(err)
	}

	for {
		head, err := mon.Recv(ctx)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("new block: %s\n", head.Hash)

		blockHashes <- head.Hash
		bw.lastBlock = head.Hash
	}
}

func (bw *BlockWatcher) GetLastBlock() tezos.BlockHash {
	return bw.lastBlock
}

type ContractWatcher struct {
	client  *rpc.Client
	address tezos.Address
}

func NewContractWatcher(client *rpc.Client, address tezos.Address) *ContractWatcher {
	return &ContractWatcher{client, address}
}

func (cw *ContractWatcher) Run(ctx context.Context, transactions chan<- *rpc.Transaction, blockHashes <-chan tezos.BlockHash) {
	defer close(transactions)
	for {
		blockHash := <-blockHashes

		operations, err := cw.client.GetBlockOperations(ctx, blockHash)

		if err != nil {
			log.Printf("err: %s\n", err)
		}

		for layer := range operations {
			for _, operation := range operations[layer] {
				for _, content := range operation.Contents {
					if transaction, ok := content.(*rpc.Transaction); ok && &transaction.Destination == &cw.address {
						transactions <- transaction
					}
				}
			}
		}
	}
}
