package tezos

import (
	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"log"
	"math/big"
	"strconv"
	"tez-ton-bridge-relay/app/everscale"
)

type QuorumContract struct {
	contract *contract.Contract
	wallet   *Wallet
	client   *rpc.Client
}

func NewQuorumContract(contract *contract.Contract, wallet *Wallet, client *rpc.Client) QuorumContract {
	return QuorumContract{contract: contract, wallet: wallet, client: client}
}

type QuorumArguments struct {
	Destination tezos.Address
	Source      tezos.Address
	Amount      tezos.N
	unwrap      UnwrapTransaction
	Branch      tezos.BlockHash
}

func (qa *QuorumArguments) WithDestination(address tezos.Address) {
	qa.Destination = address
}

func (qa *QuorumArguments) WithAmount(n tezos.N) {
	qa.Amount = n
}

func (qa *QuorumArguments) Encode() *codec.Transaction {
	return &codec.Transaction{
		Manager: codec.Manager{
			Source:       qa.Source,
			GasLimit:     10000,
			StorageLimit: 10000,
			Fee:          2000000,
		},
		Amount:      qa.Amount,
		Destination: qa.Destination,
		Parameters:  qa.Parameters(),
	}
}

func (qa *QuorumArguments) Parameters() *micheline.Parameters {
	return &micheline.Parameters{
		Entrypoint: "approveTransfer",
		Value:      micheline.NewPair(micheline.NewBytes(qa.unwrap.Address.Bytes22()), micheline.NewNat(big.NewInt(qa.unwrap.Amount.Int64()))),
	}
}

func (qa *QuorumArguments) WithSource(address tezos.Address) {
	qa.Source = address
}

type UnwrapTransaction struct {
	Address tezos.Address
	Amount  tezos.N
}

func NewUnwrapTransactionFromEvent(event *everscale.UnwrapTokenEvent) (*UnwrapTransaction, error) {
	log.Println(event)
	val, err := strconv.Atoi(event.Amount)
	if err != nil {
		return nil, err
	}

	address, err := tezos.ParseAddress(event.Addr)
	if err != nil {
		return nil, err
	}

	return &UnwrapTransaction{address, tezos.N(val)}, nil
}

func (qc *QuorumContract) SendApprove(transaction UnwrapTransaction) {
	call, err := qc.contract.Call(context.Background(), &QuorumArguments{
		Branch:      qc.wallet.bw.GetLastBlock(),
		Destination: qc.contract.Address(),
		Source:      qc.wallet.key.Address(),
		Amount:      0,
		unwrap:      transaction,
	}, &contract.CallOptions{
		Confirmations: 5,
		TTL:           120,
		Limits:        tezos.Limits{},
		MaxFee:        1000000,
		Signer:        qc.wallet,
		Observer:      qc.client.BlockObserver,
	})
	if err != nil {
		log.Printf("err: %v", err)
		return
	}
	log.Println(call)
}
