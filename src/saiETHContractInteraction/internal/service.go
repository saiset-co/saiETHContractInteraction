package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/saiset-co/saiETHContractInteraction/models"
	"github.com/saiset-co/saiService"
	"go.uber.org/zap"

	bolt "go.etcd.io/bbolt"
)

type InternalService struct {
	Handler   saiService.Handler // handlers to define in this specified microservice
	Contracts []models.Contract
	Mutex     *sync.RWMutex
	Context   *saiService.Context
	Logger    *zap.Logger
	Db        *bolt.DB
}

// global handler for registering handlers
var Service = &InternalService{
	Handler:   saiService.Handler{},
	Contracts: make([]models.Contract, 0),
	Mutex:     new(sync.RWMutex),
}

func (is *InternalService) Init() {
	fmt.Println(is.Context)
	Service.Logger = is.Context.Context.Value("logger").(*zap.Logger)

	// initializing db
	db, err := bolt.Open("db.db", 0666, nil)
	if err != nil {
		Service.Logger.Sugar().Fatalf("main - init - open db : %s", err)
	}
	is.Db = db

	is.Db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(requestBucket))
		if err != nil {
			if err == bolt.ErrBucketExists {
				return nil
			}
			Service.Logger.Sugar().Fatalf("main - init - create bucket : %s", err)
		}
		return nil
	})

	go is.handlePendingRequests(context.Background())

	Service.getInitialContracts("contracts.json")
}

// get pending requests from db, handle it and delete if request was handled successful
func (is *InternalService) handlePendingRequests(ctx context.Context) {
	sleep := is.Context.GetConfig("specific.sleep", 5).(int)
	Service.Logger.Debug("main - handle pending requests", zap.Int("timeout", sleep))
	for {
		time.Sleep(time.Duration(sleep) * time.Second)
		Service.Logger.Debug("main - handle pending requests - start")
		requests, err := is.GetPendingRequests()
		if err != nil {
			Service.Logger.Error("main - handle pending requests - getPendingRequests", zap.Error(err))
			continue
		}

		if len(requests) == 0 {
			Service.Logger.Debug("main - handle pending requests - requests not found")
			continue
		}

		for _, req := range requests {
			contract, err := Service.GetContractByName(req.Contract)
			if err != nil {
				Service.Logger.Error("main - handle pending requests - GetContractByName", zap.Error(err))
				continue
			}

			Service.Logger.Debug("main - handle pending requests - handling", zap.String("request", req.Contract), zap.String("server", contract.Server))

			ethClient, err := ethclient.Dial(contract.Server)
			if err != nil {
				Service.Logger.Error("main - handle pending requests - dial eth server", zap.String("server name", contract.Server), zap.Error(err))
				continue
			}
			// check connection (func above do not do this)
			id, err := ethClient.NetworkID(context.Background())
			if err != nil {
				Service.Logger.Error("main - handle pending requests - check eth server", zap.String("server name", contract.Server), zap.Error(err))
				continue
			}

			Service.Logger.Debug("main - handle pending requests - connection established", zap.String("network id", id.String()))

			value, input, err := Service.HandleRequest(contract, req)
			if err != nil {
				Service.Logger.Error("main - handle pending requests - HandleRequest", zap.Error(err))
				continue
			}

			response, err := Service.RawTransaction(ethClient, value, input, contract)
			if err != nil {
				Service.Logger.Error("main - handle pending requests - RawTransaction", zap.Error(err))
				continue
			}

			Service.Logger.Debug("main - handle pending requests - response", zap.String("server name", contract.Server), zap.String("response", response))

			err = is.Delete(req.DbKey)
			if err != nil {
				Service.Logger.Error("main - handle pending requests - delete", zap.Error(err))
				continue
			}

			err = is.UpdateStatus(req)
			if err != nil {
				Service.Logger.Error("main - handle pending requests - saveUpdatedRequest", zap.Error(err))
				continue
			}
		}
	}
}
