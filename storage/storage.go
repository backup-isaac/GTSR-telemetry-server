package storage

import (
	"encoding/json"
	"io/ioutil"

	client "github.com/influxdata/influxdb/client/v2"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

const tableName = "telemetry"

// Storage describes the interface with persistent storage
type Storage interface {
	// Insert inserts points into the store
	Insert(points []*datatypes.Datapoint) error
	// Close performs cleanup work
	Close() error
}

type storageImpl struct {
	client client.Client
}

// NewStorage returns an initialized Storage, backed by InfluxDB
func NewStorage() (Storage, error) {
	rawJSON, err := ioutil.ReadFile("secrets.json")
	if err != nil {
		return nil, err
	}
	secrets := make(map[string]string)
	err = json.Unmarshal(rawJSON, &secrets)
	if err != nil {
		return nil, err
	}
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     secrets["address"],
		Username: secrets["username"],
		Password: secrets["password"],
	})
	if err != nil {
		return nil, err
	}
	storage := &storageImpl{
		client: c,
	}
	return storage, nil
}

func (s *storageImpl) Insert(points []*datatypes.Datapoint) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  tableName,
		Precision: "ns",
	})
	if err != nil {
		return err
	}
	for _, point := range points {
		pt, err := client.NewPoint(point.Metric, point.Tags, map[string]interface{}{"value": point.Value})
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	return s.client.Write(bp)
}

func (s *storageImpl) Close() error {
	return s.client.Close()
}
