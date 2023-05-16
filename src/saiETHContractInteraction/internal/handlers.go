package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/saiset-co/saiETHContractInteraction/models"
	"github.com/saiset-co/saiService"
	"go.uber.org/zap"
)

const (
	contractsPath = "contracts.json"
	requestPath   = "requests.json"
)

func (is *InternalService) NewHandler() saiService.Handler {
	return saiService.Handler{
		"api": saiService.HandlerElement{
			Name:        "api",
			Description: "transact encoded transaction to contract by ABI",
			Function: func(data interface{}) (*saiService.SaiResponse, error) {
				fmt.Println("api started")
				contractData, ok := data.(map[string]interface{})
				if !ok {
					Service.Logger.Sugar().Debugf("handling connect method, wrong type, current type : %+v", reflect.TypeOf(data))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, errors.New("wrong type of incoming data")
				}

				b, err := json.Marshal(contractData)
				if err != nil {
					Service.Logger.Error("handlers - api - marshal incoming data", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				req := models.EthRequest{}
				err = json.Unmarshal(b, &req)
				if err != nil {
					Service.Logger.Error("handlers - api - unmarshal data to struct", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				contract, err := Service.GetContractByName(req.Contract)
				if err != nil {
					Service.Logger.Error("handlers - api - GetContractByName", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				ethClient, err := ethclient.Dial(contract.Server)
				if err != nil {
					Service.Logger.Error("handlers - api - dial eth server", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err

				}

				// check connection (func above do not do this)
				id, err := ethClient.NetworkID(context.Background())
				if err != nil {
					Service.Logger.Error("handlers - api - check eth server", zap.Error(err))
					Service.Logger.Debug("handlers - api - check connection - error - saving request to db")

					uid, err := is.Save(&req, b)
					if err != nil {
						Service.Logger.Error("handlers - api - save request", zap.Error(err))
						return &saiService.SaiResponse{
							StatusCode: http.StatusBadRequest,
						}, err
					}

					return &saiService.SaiResponse{
						StatusCode: http.StatusOK,
						Data:       uid,
					}, nil
				}

				Service.Logger.Debug("handlers - api - connection established", zap.String("network id", id.String()))

				value, input, err := Service.HandleRequest(contract, &req)
				if err != nil {
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				response, err := Service.RawTransaction(ethClient, value, input, contract)
				if err != nil {
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				return &saiService.SaiResponse{
					StatusCode: http.StatusOK,
					Data:       response,
				}, nil
			},
		},

		"add": saiService.HandlerElement{
			Name:        "add",
			Description: "add contract to contracts",
			Function: func(data interface{}) (*saiService.SaiResponse, error) {
				contractData, ok := data.(map[string]interface{})
				if !ok {
					Service.Logger.Sugar().Debugf("handlers - add - wrong data type, current type : %+v", data)
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, errors.New("wrong type of incoming data")
				}

				b, err := json.Marshal(contractData)
				if err != nil {
					Service.Logger.Error("api - add - marshal incoming data", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				contracts := models.Contracts{}
				err = json.Unmarshal(b, &contracts)
				if err != nil {
					Service.Logger.Error("handlers - add - unmarshal data to struct", zap.Error(err))
					return nil, err
				}

				// validate all incoming contracts
				validatedContracts := make([]models.Contract, 0)
				for _, contract := range contracts.Contracts {
					err = contract.Validate()
					if err != nil {
						Service.Logger.Error("handlers - add - validate incoming contracts", zap.Any("contract", contract), zap.Error(err))
						continue
					}
					validatedContracts = append(validatedContracts, contract)
				}

				// check if incoming contracts already exists
				Service.Mutex.RLock()
				checkedContracts := Service.filterUniqueContracts(validatedContracts)
				Service.Mutex.RUnlock()

				Service.Mutex.Lock()
				Service.Contracts = append(Service.Contracts, checkedContracts...)
				Service.Mutex.Unlock()

				//	Service.Logger.Sugar().Debugf("ACTUAL CONTRACTS : %+v", Service.Contracts)

				err = Service.RewriteContractsConfig(contractsPath)
				if err != nil {
					Service.Logger.Error("handlers - add - rewrite contracts file", zap.Error(err))
					return nil, err
				}
				return &saiService.SaiResponse{
					StatusCode: http.StatusOK,
					Data:       "OK",
				}, nil

			},
		},

		"delete": saiService.HandlerElement{
			Name:        "delete",
			Description: "delete contract by name",
			Function: func(data interface{}) (*saiService.SaiResponse, error) {
				deleteData, ok := data.(map[string]interface{})
				if !ok {
					Service.Logger.Sugar().Debugf("handlers - delete - wrong data type, current type : %+v", data)
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, errors.New("wrong type of incoming data")
				}

				b, err := json.Marshal(deleteData)
				if err != nil {
					Service.Logger.Error("api - delete - marshal incoming data", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				deleteContractName := models.DeleteData{}
				err = json.Unmarshal(b, &deleteContractName)
				if err != nil {
					Service.Logger.Error("handlers - delete - unmarshal data to struct", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				Service.Mutex.Lock()
				Service.DeleteContracts(&deleteContractName)
				Service.Mutex.Unlock()

				Service.Logger.Sugar().Debugf("CONTRACTS AFTER DELETION : %+v", Service.Contracts)

				err = Service.RewriteContractsConfig(contractsPath)
				if err != nil {
					Service.Logger.Error("handlers - delete - rewrite contracts file", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusInternalServerError,
					}, err
				}
				return &saiService.SaiResponse{
					StatusCode: http.StatusOK,
					Data:       "OK",
				}, nil

			},
		},
		"checkStatus": saiService.HandlerElement{
			Name:        "checkStatus",
			Description: "check status of pending request",
			Function: func(data interface{}) (*saiService.SaiResponse, error) {
				checkData, ok := data.(map[string]interface{})
				if !ok {
					Service.Logger.Sugar().Debugf("handlers - checkStatus - wrong data type, current type : %+v", data)
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, errors.New("wrong type of incoming data")
				}
				b, err := json.Marshal(checkData)
				if err != nil {
					Service.Logger.Error("api - checkStatus - marshal incoming data", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				req := models.CheckStatusRequest{}
				err = json.Unmarshal(b, &req)
				if err != nil {
					Service.Logger.Error("handlers - checkStatus - unmarshal data to struct", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				err = req.Validate()
				if err != nil {
					Service.Logger.Error("handlers - checkStatus - validate", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				status, err := is.Get(req.ID)
				if err != nil {
					Service.Logger.Error("handlers - checkStatus - db.Get", zap.Error(err))
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}

				return &saiService.SaiResponse{
					StatusCode: http.StatusOK,
					Data:       status,
				}, err

			},
		},
		"getAll": saiService.HandlerElement{ //for testing purposes
			Name:        "getAll",
			Description: "get all keys",
			Function: func(data interface{}) (*saiService.SaiResponse, error) {
				req, err := is.GetPendingRequests()
				if err != nil {
					return &saiService.SaiResponse{
						StatusCode: http.StatusBadRequest,
					}, err
				}
				return &saiService.SaiResponse{
					StatusCode: http.StatusOK,
					Data:       req,
				}, err
			},
		},
	}
}
