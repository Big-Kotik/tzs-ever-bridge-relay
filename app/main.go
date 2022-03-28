package main

import (
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"encoding/json"
	shell "github.com/ipfs/go-ipfs-api"
	goton "github.com/move-ton/ton-client-go"
	"github.com/move-ton/ton-client-go/domain"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"strings"
	ever "tez-ton-bridge-relay/app/everscale"
	"tez-ton-bridge-relay/app/ipfs"
	tz "tez-ton-bridge-relay/app/tezos"
)

type Config struct {
	TezosConfig tz.Config   `yaml:"tezosConfig"`
	EverConfig  ever.Config `yaml:"everscaleConfig"`
	IPFSConfig  ipfs.Config `yaml:"ipfsConfig"`
}

func (c *Config) GetConfigFromFile(file string) error {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return err
	}

	return nil
}

type Tezos struct {
	Client             *rpc.Client
	Wallet             *tz.Wallet
	QuorumContract     *tz.QuorumContract
	DepositTransaction chan *rpc.Transaction
}

func RunTezosSide(config *tz.Config) (*Tezos, error) {
	ctx := context.TODO()

	depositAddress := tezos.MustParseAddress(config.Contracts.DepositAddress)

	c, err := rpc.NewClient(config.Server.TezosURL, nil)

	if err != nil {
		return nil, err
	}

	if err = c.Init(ctx); err != nil {
		return nil, err
	}
	go c.Listen()

	headBlock, err := c.GetHeadBlock(ctx)
	if err != nil {
		return nil, err
	}

	blockWatcher := tz.NewBlockWatcher(c, headBlock.Hash)
	contractWatcher := tz.NewContractWatcher(c, depositAddress)

	blocks := make(chan tezos.BlockHash)
	transactions := make(chan *rpc.Transaction)

	go blockWatcher.Run(ctx, blocks)
	go contractWatcher.Run(ctx, transactions, blocks)

	pk, err := tezos.ParsePrivateKey(config.Wallet.PrivateKey)

	if err != nil {
		return nil, err
	}

	w := tz.NewWallet(c, blockWatcher, pk)

	quorumAddress, err := tezos.ParseAddress(config.Contracts.QuorumAddress)

	if err != nil {
		return nil, err
	}

	contr := tz.NewQuorumContract(contract.NewContract(quorumAddress, c), w, c)

	return &Tezos{
		QuorumContract:     &contr,
		Client:             c,
		Wallet:             w,
		DepositTransaction: transactions,
	}, nil
}

type Everscale struct {
	Ton    *goton.Ton
	Events <-chan *domain.DecodedMessageBody
}

func RunEverSide(config *ever.Config) (*Everscale, error) {
	ton, err := goton.NewTon("", config.Servers)

	if err != nil {
		return nil, err
	}

	file, err := os.Open(config.Contracts.DepositContract.ABI)

	if err != nil {
		return nil, err
	}

	byteAbi, err := ioutil.ReadAll(file)

	if err != nil {
		return nil, err
	}

	nn := &domain.AbiContract{}
	if err = json.Unmarshal(byteAbi, &nn); err != nil {
		return nil, err
	}

	testAbi := domain.NewAbiContract(nn)

	c := ever.NewEventWatcher(config.Contracts.DepositContract.Address, ton, testAbi)

	cs := make(chan *domain.DecodedMessageBody)
	go c.RunWatcher(cs, func(message *domain.DecodedMessageBody) bool {
		return message.Name == "UnwrapTokenEvent"
	})

	return &Everscale{Events: cs, Ton: ton}, err
}

func main() {
	config := &Config{}
	if err := config.GetConfigFromFile("config.yaml"); err != nil {
		log.Printf("err: %s\n", err)
		return
	}

	tezosClient, err := RunTezosSide(&config.TezosConfig)

	if err != nil {
		log.Printf("err: %s\n", err)
		return
	}

	everscale, err := RunEverSide(&config.EverConfig)

	if err != nil {
		log.Printf("err: %s\n", err)
		return
	}

	defer everscale.Ton.Client.Destroy()

	sh := shell.NewShell(config.IPFSConfig.Server)

	for {
		select {
		case transaction := <-tezosClient.DepositTransaction:
			log.Printf("Transaction: %s", tz.ParseFromTransaction(transaction).String())
			cid, err := sh.Add(strings.NewReader(tz.ParseFromTransaction(transaction).String()))
			if err != nil {
				log.Printf("err: %s\n", err)
				continue
			}
			log.Printf("IPFS Hash %s\n", cid)
		case event := <-everscale.Events:
			var msg ever.UnwrapTokenEvent
			err := json.Unmarshal(event.Value, &msg)
			if err != nil {
				// TODO: Should reject transaction
				log.Printf("err: %s\n", err)
				continue
			}

			transaction, err := tz.NewUnwrapTransactionFromEvent(&msg)

			if err != nil {
				// TODO: Should reject transaction
				log.Printf("err: %s\n", err)
				continue
			}

			go tezosClient.QuorumContract.SendApprove(*transaction)
		}
	}

}
