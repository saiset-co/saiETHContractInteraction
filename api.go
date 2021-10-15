package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/common/log"
	"go.mongodb.org/mongo-driver/bson"
	"math/big"
	"net/http"
	"strings"
)

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

type Arg struct {
	Type 	string		`json:"type"`
	Value 	interface{} `json:"value"`
}

type JsonData struct {
	Function	string	`json:"function"`
	Args     	[]Arg	`json:"args"`
}

func hasAccess(r *http.Request) bool  {
	headers := r.Header
	token, ok := headers["Token"]

	if !ok {
		return false
	}

	if len(token) > 0 && token[0] == config.Token {
		return true
	}

	return false
}

func api(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)

	if !hasAccess(r) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		log.Warn("Unauthorized access")
		message, _ := json.Marshal(bson.M{"message": "Unauthorized access"})
		_, _ = w.Write(message)
		return
	}

	err := r.ParseForm()

	if err != nil {
		log.Fatalf("Could not parse form: %v", err)
	}

	abiEl, err := abi.JSON(strings.NewReader(config.Contract.ABI))

	if err != nil {
		log.Fatalf("Could not read ABI: %v", err)
	}

	decoder := json.NewDecoder(r.Body)
	var result JsonData
	var args []interface{}
	err = decoder.Decode(&result)

	if err != nil {
		log.Fatalf("Wrong JSON: %v", err)
	}

	client, err := ethclient.Dial(config.Geth)

	if err != nil {
		log.Fatalf("Failed to connect to the ethereum server: %v", err)
	}

	for _, v := range result.Args {
		arg := v.Value

		if v.Type == "address" {
			arg = common.HexToAddress(v.Value.(string))
		}

		if v.Type == "[]string" {
			t := v.Value.([]interface{})
			s := make([]string, len(t))
			for i, a := range t {
				s[i] = fmt.Sprint(a)
			}

			arg = s
		}

		args = append(args, arg)
	}

	input, _ := abiEl.Pack(result.Function, args...)

	response := rawTransaction(client, big.NewInt(0), input)
	_, _ = w.Write([]byte(response))

}
