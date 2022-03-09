package main

import (
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"encoding/json"
	"fmt"
	goton "github.com/move-ton/ton-client-go"
	"github.com/move-ton/ton-client-go/domain"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	ever "tez-ton-bridge-relay/app/everscale"
	tz "tez-ton-bridge-relay/app/tezos"
)

type Config struct {
	TezosConfig tz.TezosConfig       `yaml:"tezosConfig"`
	EverConfig  ever.EverscaleConfig `yaml:"everscaleConfig"`
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

func RunTezosSide(config *tz.TezosConfig) (*Tezos, error) {
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

	blockWatcher := tz.NewBlockWatcher(c)
	contractWatcher := tz.NewContractWatcher(c, blockWatcher, depositAddress)

	go blockWatcher.Run(ctx)
	go contractWatcher.Run(ctx)

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
		DepositTransaction: contractWatcher.Transactions,
	}, nil
}

type Everscale struct {
	Ton    *goton.Ton
	Events <-chan *domain.DecodedMessageBody
}

func RunEverSide(config *ever.EverscaleConfig) (*Everscale, error) {
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
	log.Println(tezosClient.Client.ChainId)

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

	for {
		select {
		case transaction := <-tezosClient.DepositTransaction:
			fmt.Println(tz.ParseFromTransaction(transaction))
		case event := <-everscale.Events:
			var msg ever.UnwrapTokenEvent
			err := json.Unmarshal(event.Value, &msg)
			if err != nil {
				log.Printf("err: %s\n", err)
				continue
			}

			transaction, err := tz.NewUnwrapTransactionFromEvent(&msg)

			if err != nil {
				log.Printf("err: %s\n", err)
				continue
			}

			go tezosClient.QuorumContract.SendApprove(*transaction)
		}
	}

}
