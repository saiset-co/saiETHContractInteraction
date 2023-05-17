package models

import (
	valid "github.com/asaskevich/govalidator"
)

type Parameter struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type EthRequest struct {
	Contract    string       `json:"contract"`
	Method      string       `json:"method"`
	Value       string       `json:"value"`
	Params      []*Parameter `json:"params"`
	IsProcessed bool         `json:"is_processed"`
	DbKey       string       `json:"db_key,omitempty"`
}

type Contract struct {
	Name     string `json:"name" valid:",required"`
	Server   string `json:"server" valid:",required"`
	ABI      string `json:"abi" valid:",required"`
	Address  string `json:"address" valid:",required"`
	Private  string `json:"private" valid:",required"`
	GasLimit uint64 `json:"gas_limit" valid:",required"`
}

// Validate contract
func (m *Contract) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}

type Contracts struct {
	Contracts []Contract `json:"contracts"`
}
type DeleteData struct {
	Names []string `json:"names"`
}

type CheckStatusRequest struct {
	ID string `json:"id" valid:",required"`
}

// Validate contract
func (m *CheckStatusRequest) Validate() error {
	_, err := valid.ValidateStruct(m)
	return err
}
