package internal

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/saiset-co/saiETHContractInteraction/models"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

var (
	requestBucket = []byte("request_bucket")
)

// save requests in boltdb
func (is *InternalService) Save(data []byte) (string, error) {
	uid := uuid.New()
	err := is.Db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(requestBucket)
		err := b.Put([]byte(uid.String()), data)
		if err != nil {
			return fmt.Errorf("db - db.Put : %w", err)
		}
		return nil
	})
	return uid.String(), err
}

// get pending requests from db
func (is *InternalService) GetPendingRequests() ([]*models.EthRequest, error) {
	requests := make([]*models.EthRequest, 0)
	is.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(requestBucket)

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Println(string(k))
			req := models.EthRequest{}
			err := json.Unmarshal(v, &req)
			if err != nil {
				Service.Logger.Error("db - GetPendingRequests - unmarshal", zap.String("request", string(k)), zap.Error(err))
				continue
			}
			req.DbKey = string(k)
			requests = append(requests, &req)
		}

		return nil
	})
	return requests, nil
}

// get request by id
func (is *InternalService) Get(id string) (bool, error) {
	var status bool
	err := is.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(requestBucket)
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("request with id=%s was not found", string(id))
		}
		req := models.EthRequest{}
		err := json.Unmarshal(data, &req)
		if err != nil {
			Service.Logger.Error("db - Get - unmarshal", zap.Error(err))
			Service.Logger.Debug("db - Get - body to unmarshal", zap.String("body", string(data)))
			return err
		}
		status = req.IsProcessed
		return nil
	})
	return status, err
}

// delete request by id
func (is *InternalService) Delete(id string) error {
	err := is.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(requestBucket)
		err := b.Delete([]byte(id))
		return err
	})
	return err
}

// save requests in boltdb
func (is *InternalService) UpdateStatus(req *models.EthRequest) error {
	err := is.Db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(requestBucket)
		err := b.Delete([]byte(req.DbKey))
		if err != nil {
			return fmt.Errorf("db - updateStatus - delete : %w", err)
		}

		req.IsProcessed = true

		data, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("db - updateStatus - marshal : %w", err)
		}
		b.Put([]byte(req.DbKey), data)
		return nil
	})
	return err
}
