package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/common/log"
	"github.com/tkanos/gonfig"
	"math/big"
	"net/http"
	"time"
)

type configuration struct {
	HttpServer struct {
		Host string
		Port string
	}
	Storage struct {
		Url  string
		Auth struct {
			Email    string
			Password string
		}
	}
	Token string
	Geth  string
	GasLimit int
	Contract struct {
		ABI string
		Address string
		Private string
	}
	Crypto []string
}

var config configuration

func main() {
	configErr := gonfig.GetConf("config.json", &config)

	if configErr != nil {
		fmt.Println("Config missed!! ")
		panic(configErr)
	}

	fmt.Println("Server start: http://" + config.HttpServer.Host + ":" + config.HttpServer.Port)

	http.HandleFunc("/", api)

	serverErr := http.ListenAndServe(config.HttpServer.Host+":"+config.HttpServer.Port, nil)

	if serverErr != nil {
		fmt.Println("Server error: ", serverErr)
	}
}

func rawTransaction(client *ethclient.Client, value *big.Int, data []byte) string {
	d := time.Now().Add(5000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	privateKey, err := crypto.HexToECDSA(config.Contract.Private)
	if err != nil {
		log.Info(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Info(err)
	}

	toAddress := common.HexToAddress(config.Contract.Address)
	gasLimit := uint64(config.GasLimit)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Info(err)
	}

	log.Infof("Gas price: %v", gasPrice)

	if !ok {
		log.Info("Error converting amount")
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Info(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Info(err)
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Info(err)
	}

	return signedTx.Hash().String()
}
