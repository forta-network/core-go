package main

import (
	"context"
	"math/big"

	"github.com/forta-network/core-go/etherclient"
	"github.com/forta-network/core-go/utils"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Env struct {
	JSONRPCURL string `envconfig:"JSON_RPC_URL"`
}

var env Env

func main() {
	envconfig.MustProcess("", &env)

	ethClient, err := etherclient.DialContext(context.Background(), env.JSONRPCURL)
	utils.FatalIfError(err)

	blockNum, err := ethClient.BlockNumber(context.Background())
	utils.FatalIfError(err)

	for {
		txs, err := ethClient.GetBlockTransactions(context.Background(), big.NewInt(0).SetUint64(blockNum))
		utils.FatalIfError(err)

		log := logrus.WithField("block", blockNum)
		log.WithField("txCount", len(txs)).Info("got transactions")
		for i, tx := range txs {
			logrus.WithField("index", i).WithField("txType", tx.Type).WithField("hash", tx.Hash).Info("tx")
		}

		blockNum++
	}
}
