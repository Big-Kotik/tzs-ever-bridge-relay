package main

import (
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"encoding/hex"
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

func RunTezosSide(config *tz.TezosConfig) (chan *rpc.Transaction, *tz.Wallet, error) {
	ctx := context.TODO()

	depositAddress := tezos.MustParseAddress(config.Contracts.DepositAddress)

	c, err := rpc.NewClient(config.Server.TezosURL, nil)
	if err != nil {
		return nil, nil, err
	}

	blockWatcher := tz.NewBlockWatcher(c)
	contractWatcher := tz.NewContractWatcher(c, blockWatcher, depositAddress)

	go blockWatcher.Run(ctx)
	go contractWatcher.Run(ctx)

	pk, _ := tezos.ParsePrivateKey(config.Wallet.PrivateKey)

	w := tz.NewWallet(c, blockWatcher, pk)

	return contractWatcher.Transactions, w, nil
}

func RunEverSide(config *ever.EverscaleConfig) (*goton.Ton, <-chan *domain.DecodedMessageBody) {
	//log.Println(config.Servers)
	ton, err := goton.NewTon("", config.Servers)

	if err != nil {
		log.Fatalln(err)
	}

	//filter, _ := json.Marshal(json.RawMessage(`{"account_addr":{"eq":"0:15e76544085a1cd233f2caf115a08d137b6eb2301c5075cd6cb785ab48ebd85a"}}`))

	if err != nil {
		log.Println(err)
		return nil, nil
	}

	file, _ := os.Open(config.Contracts.DepositContract.ABI)
	byteAbi, _ := ioutil.ReadAll(file)

	nn := &domain.AbiContract{}
	json.Unmarshal(byteAbi, &nn)

	testAbi := domain.NewAbiContract(nn)

	c := ever.New(config.Contracts.DepositContract.Address, ton, testAbi)

	cs := make(chan *domain.DecodedMessageBody)
	go c.RunWatcher(cs)

	return ton, cs
}

func main() {
	config := &Config{}
	if err := config.GetConfigFromFile("config.yaml"); err != nil {
		log.Println(err)
		return
	}

	tezosTransChan, w, err := RunTezosSide(&config.TezosConfig)

	if err != nil {
		log.Println(err)
		return
	}

	ton, collection := RunEverSide(&config.EverConfig)

	defer ton.Client.Destroy()

	for {
		select {
		case transaction := <-tezosTransChan:
			fmt.Println(tz.ParseFromTransaction(transaction))
		case event := <-collection:
			if event.Name == "UnwrapTokenEvent" {
				var msg ever.UnwrapTokenEvent
				err := json.Unmarshal(event.Value, &msg)
				if err != nil {
					log.Println(err)
				}

				h, err := hex.DecodeString(msg.Amount[2:])
				if err != nil {
					log.Println(err)
				}
				num := uint64(0)
				for i, val := range h[24:] {
					num += uint64(val) << (8 * (7 - i))
				}

				w.SendUnwrapTransaction(context.TODO(), tz.UnwrapTransaction{msg.Addr, num})
			}
		}
	}
}
