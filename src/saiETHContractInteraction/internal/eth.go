package internal

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/saiset-co/saiETHContractInteraction/models"
	"go.uber.org/zap"
)

var mux sync.Mutex
var nonceList = map[string]map[uint64]bool{}

type resTX struct {
	Transaction *types.Transaction
	Status      string `json:"status"`
	Result      string `json:"result"`
}

func response(tx *types.Transaction, res string) (resTX, error) {
	result := resTX{
		Transaction: tx,
		Status:      "error",
		Result:      res,
	}
	resultS, _ := json.Marshal(result)
	err := errors.New(string(resultS))

	return result, err
}

func (is *InternalService) RawTransaction(client *ethclient.Client, value *big.Int, data []byte, contract *models.Contract) (string, error) {
	d := time.Now().Add(5000 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	privateKey, err := crypto.HexToECDSA(contract.Private)
	if err != nil {
		is.Logger.Error("handlers - api - RawTransaction - HexToECDSA", zap.Error(err))
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		is.Logger.Error("handlers - api - RawTransaction - cast publicKey to ecdsa", zap.Error(err))
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	toAddress := common.HexToAddress(contract.Address)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		is.Logger.Error("handlers - api - RawTransaction - get suggested gas price", zap.Error(err))
		return "", err
	}

	is.Logger.Sugar().Debugf("GAS PRICE : %v", gasPrice)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    value,
		Gas:      contract.GasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		res, err := response(tx, err.Error())
		is.Logger.Error("handlers - api - RawTransaction - get networkID", zap.Any("TX", res))
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		res, err := response(tx, err.Error())
		is.Logger.Error("handlers - api - RawTransaction - signTx", zap.Any("TX", res))
		return "", err
	}

	mux.Lock()
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		res, err := response(tx, err.Error())
		is.Logger.Error("handlers - api - RawTransaction - sendTx", zap.Any("TX", res))
		return "", err
	}

	res, _ := response(tx, signedTx.Hash().String())
	resS, _ := json.Marshal(res)

	for {
		resTx, isPending, err := client.TransactionByHash(ctx, signedTx.Hash())
		if err != nil {
			res, err := response(tx, err.Error())
			is.Logger.Error("handlers - api - RawTransaction - sendTx", zap.Any("TX", res))
			return "", err
		}

		if resTx == nil {
			is.Logger.Debug("handlers - api - RawTransaction - sendTx - tx was not created", zap.Any("TX", res))
			mux.Unlock()
			goto done
		} else if resTx != nil && !isPending {
			is.Logger.Debug("handlers - api - RawTransaction - sendTx - tx done", zap.Any("TX", res))
			mux.Unlock()
			goto done
		}

		is.Logger.Debug("handlers - api - RawTransaction - sendTx - pending, sleep 2 sec")
		time.Sleep(2 * time.Second)
	}

done:
	return string(resS), nil
}

func (is *InternalService) HandleRequest(contract *models.Contract, req *models.EthRequest) (value *big.Int, input []byte, err error) {
	abiEl, err := abi.JSON(strings.NewReader(contract.ABI))
	if err != nil {
		log.Fatalf("Could not read ABI: %v", err)
	}

	var args []interface{}
	for _, v := range req.Params {
		arg := v.Value

		switch arg.(type) {
		case float64:
			return &big.Int{}, nil, errors.New("handlers - api - wrong value format, please use strings always: 'value': '1'")
		case []float64:
			Service.Logger.Error("handlers - api - wrong value format, please use strings always: 'value': ['1']")
			return &big.Int{}, nil, errors.New("handlers - api - wrong value format, please use strings always: 'value': ['1']")
		}

		if v.Type == "address" {
			arg = common.HexToAddress(v.Value.(string))
		}

		if v.Type == "uint16" {
			num, err := strconv.ParseUint(v.Value.(string), 10, 16)
			if err != nil {
				Service.Logger.Error("handlers - api - can't convert to uint16")
				return &big.Int{}, nil, errors.New("handlers - api - can't convert to uint16")
			}
			arg = uint16(num)
		}

		if v.Type == "uint8" {
			num, err := strconv.ParseUint(v.Value.(string), 10, 8)
			if err != nil {
				Service.Logger.Error("handlers - api - can't convert to uint8")
				return &big.Int{}, nil, errors.New("handlers - api - can't convert to uint8")
			}
			arg = uint8(num)
		}
		var (
			ok bool
		)

		if v.Type == "uint256" {
			arg, ok = new(big.Int).SetString(v.Value.(string), 10)
			if !ok {
				Service.Logger.Error("handlers - api - can't convert to bigInt")
				return &big.Int{}, nil, errors.New("handlers - api - can't convert to bigInt")
			}
		}

		if v.Type == "address[]" {
			t := v.Value.([]interface{})
			s := make([]common.Address, len(t))
			for i, a := range t {
				s[i] = common.HexToAddress(a.(string))
			}
			arg = s
		}

		if v.Type == "string[]" {
			t := v.Value.([]interface{})
			s := make([]string, len(t))
			for i, a := range t {
				s[i] = fmt.Sprint(a)
			}
			arg = s
		}

		if v.Type == "uint256[]" {
			t := v.Value.([]interface{})
			s := make([]*big.Int, len(t))
			for i, a := range t {
				s[i], ok = new(big.Int).SetString(a.(string), 10)
				if !ok {
					Service.Logger.Error("handlers - api - can't convert to bigInt uint256[]")
					return &big.Int{}, nil, errors.New("handlers - api - can't convert to bigInt uint256[]")
				}
			}
			arg = s
		}

		if v.Type == "uint16[]" {
			t := v.Value.([]interface{})
			s := make([]uint16, len(t))
			for i, a := range t {
				num, err := strconv.ParseUint(a.(string), 10, 16)
				if err != nil {
					Service.Logger.Error("handlers - api - can't convert to uint16 uint16[]")
					return &big.Int{}, nil, errors.New("handlers - api - can't convert to uint16 uint16[]")
				}
				s[i] = uint16(num)
			}
			arg = s
		}

		if v.Type == "uint8[]" {
			t := v.Value.([]interface{})
			s := make([]uint8, len(t))
			for i, a := range t {
				num, err := strconv.ParseUint(a.(string), 10, 8)
				if err != nil {
					Service.Logger.Error("handlers - api - can't convert to uint8 uint8[]")
					return &big.Int{}, nil, errors.New("handlers - api - can't convert to uint8 uint8[]")
				}
				s[i] = uint8(num)
			}
			arg = s
		}

		args = append(args, arg)

		Service.Logger.Info("handlers - api", zap.Any("args", args))
	}

	input, err = abiEl.Pack(req.Method, args...)

	if err != nil {
		Service.Logger.Error("handlers - api - pack eth server", zap.Error(err))
		return &big.Int{}, nil, err
	}

	var (
		ok bool
	)

	if req.Value != "" {
		value, ok = new(big.Int).SetString(req.Value, 10)
		if !ok {
			Service.Logger.Error("handlers - api - can't convert value to bigInt")
			return &big.Int{}, nil, errors.New("handlers - api - can't convert value `to bigInt")
		}
	}
	return value, input, nil
}
