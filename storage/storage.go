package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.gatech.edu/GTSR/telemetry-server/datatypes"
)

const tableName = "telemetry"

// Storage describes the interface with persistent storage
type Storage interface {
	// Insert inserts points into the store
	Insert(points []*datatypes.Datapoint) error
	// DeleteMetric deletes a metric from the store
	DeleteMetric(metric string) error
	// Select all entries for specified metric
	SelectMetric(metric string) ([]*datatypes.Datapoint, error)
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
		fields := map[string]interface{}{"value": point.Value}
		pt, err := getPoint(point.Metric, point.Tags, fields, point.Time)
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	return s.client.Write(bp)
}

func getPoint(metric string, tags map[string]string, fields map[string]interface{}, time time.Time) (*client.Point, error) {
	if time.IsZero() {
		return client.NewPoint(metric, tags, fields)
	}
	return client.NewPoint(metric, tags, fields, time)
}

func (s *storageImpl) DeleteMetric(metric string) error {
	response, err := s.client.Query(client.Query{
		Command:  fmt.Sprintf("DROP MEASUREMENT \"%s\"", metric),
		Database: tableName,
	})
	if err != nil {
		return err
	}
	return response.Error()
}

func (s *storageImpl) SelectMetric(metric string) ([]*datatypes.Datapoint, error) {
	response, err := s.client.Query(client.Query{
		Command:  fmt.Sprintf("SELECT * FROM \"%s\"", metric),
		Database: tableName,
	})
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}
	if len(response.Results) == 0 || len(response.Results[0].Series) == 0 {
		return make([]*datatypes.Datapoint, 0), nil
	}
	var timeColumn, valueColumn int
	for i, columnName := range response.Results[0].Series[0].Columns {
		if columnName == "time" {
			timeColumn = i
		} else if columnName == "value" {
			valueColumn = i
		}
	}
	values := response.Results[0].Series[0].Values
	results := make([]*datatypes.Datapoint, len(values))
	for i, value := range values {
		timestamp, err := time.Parse(time.RFC3339Nano, value[timeColumn].(string))
		if err != nil {
			return nil, err
		}
		results[i] = &datatypes.Datapoint{
			Metric: metric,
			Value:  value[valueColumn],
			Tags:   response.Results[0].Series[0].Tags,
			Time:   timestamp,
		}
	}
	return results, nil
}

func (s *storageImpl) Close() error {
	return s.client.Close()
}
