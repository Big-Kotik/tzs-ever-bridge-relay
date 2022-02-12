package main

import (
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	tz "tez-ton-bridge-relay/app/tezos"
	"time"
)

type Config struct {
	Server struct {
		TezosURL string `yaml:"tezosURL"`
	} `yaml:"server"`
	TezosWallet struct {
		PrivateKey string `yaml:"privateKey"`
	} `yaml:"tezosWallet"`
	TezosContracts struct {
		QuorumAddress  string `yaml:"quorumAddress"`
		DepositAddress string `yaml:"depositAddress"`
	} `yaml:"tezosContracts"`
}

func (c *Config) GetConfigFromFile(file string) error {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	log.Println(string(yamlFile))
	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return err
	}
	return nil
}

func main() {
	config := &Config{}
	if err := config.GetConfigFromFile("config.yaml"); err != nil {
		log.Println(err)
		return
	}

	ctx := context.TODO()

	depositAddress := tezos.MustParseAddress(config.TezosContracts.DepositAddress)
	c, _ := rpc.NewClient(config.Server.TezosURL, nil)

	blockWatcher := tz.NewBlockWatcher(c)
	contractWatcher := tz.NewContractWatcher(c, blockWatcher, depositAddress)

	go blockWatcher.Run(ctx)
	go contractWatcher.Run(ctx)

	pk, _ := tezos.ParsePrivateKey(config.TezosWallet.PrivateKey)

	w := tz.NewWallet(c, blockWatcher, pk)

	//TODO: Change to get HeadBlock
	time.Sleep(1000 * time.Millisecond)
	w.SendTransaction(ctx, tezos.MustParseAddress(config.TezosContracts.QuorumAddress))

	for {
		transaction := <-contractWatcher.Transactions

		fmt.Println(tz.ParseFromTransaction(transaction))
	}
}
