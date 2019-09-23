package storage

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"server/datatypes"

	client "github.com/influxdata/influxdb/client/v2"
)

const tableName = "telemetry"

var metricRegex = regexp.MustCompile("\\A[a-zA-Z0-9_-]*\\z")

// ValidMetric returns whether the metric name is valid
func ValidMetric(metric string) bool {
	return metricRegex.MatchString(metric)
}

// Storage describes the interface with persistent storage
type Storage struct {
	client client.Client
}

// NewStorage returns an initialized Storage, backed by InfluxDB
func NewStorage() (*Storage, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: "http://influxdb:8086",
	})
	if err != nil {
		return nil, err
	}
	storage := &Storage{
		client: c,
	}
	return storage, nil
}

// Insert inserts points into the store
func (s *Storage) Insert(points []*datatypes.Datapoint) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  tableName,
		Precision: "ns",
	})
	if err != nil {
		return err
	}
	for _, point := range points {
		if !ValidMetric(point.Metric) {
			return metricError(point.Metric)
		}
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

// DeleteMetric deletes a metric from the store
func (s *Storage) DeleteMetric(metric string) error {
	if !ValidMetric(metric) {
		return metricError(metric)
	}
	response, err := s.client.Query(client.Query{
		Command:  fmt.Sprintf("DROP MEASUREMENT %s", metric),
		Database: tableName,
	})
	if err != nil {
		return err
	}
	return response.Error()
}

// SelectMetric selects all entries for specified metric
func (s *Storage) SelectMetric(metric string) ([]*datatypes.Datapoint, error) {
	if !ValidMetric(metric) {
		return nil, metricError(metric)
	}
	response, err := s.client.Query(client.Query{
		Command:  fmt.Sprintf("SELECT * FROM %s", metric),
		Database: tableName,
	})
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}
	return getDatapoints(metric, response)
}

// SelectMetricTimeRange selects entries for metric within specified time range
func (s *Storage) SelectMetricTimeRange(metric string, start time.Time, end time.Time) ([]*datatypes.Datapoint, error) {
	if !ValidMetric(metric) {
		return nil, metricError(metric)
	}
	response, err := s.client.Query(client.Query{
		Command: fmt.Sprintf("SELECT * FROM %s WHERE time >= '%s' AND time <= '%s'",
			metric, start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano)),
		Database: tableName,
	})
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}
	return getDatapoints(metric, response)
}

func getDatapoints(metric string, response *client.Response) ([]*datatypes.Datapoint, error) {
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
		val, err := strconv.ParseFloat(string(value[valueColumn].(json.Number)), 64)
		if err != nil {
			return nil, err
		}
		results[i] = &datatypes.Datapoint{
			Metric: metric,
			Value:  val,
			Tags:   response.Results[0].Series[0].Tags,
			Time:   timestamp,
		}
	}
	return results, nil
}

// ListMetrics lists all of the metrics in the table
func (s *Storage) ListMetrics() ([]string, error) {
	response, err := s.client.Query(client.Query{
		Command:  "SHOW MEASUREMENTS",
		Database: tableName,
	})
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}
	values := response.Results[0].Series[0].Values
	metrics := make([]string, len(values))
	for i, value := range values {
		metrics[i] = value[0].(string)
	}
	return metrics, nil
}

// Latest returns the most recent datapoint for the given metric
func (s *Storage) Latest(metric string) (*datatypes.Datapoint, error) {
	if !ValidMetric(metric) {
		return nil, metricError(metric)
	}
	response, err := s.client.Query(client.Query{
		Command:  fmt.Sprintf("SELECT * FROM %s ORDER BY DESC LIMIT 1", metric),
		Database: tableName,
	})
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}
	points, err := getDatapoints(metric, response)
	if err != nil {
		return nil, err
	}
	if len(points) != 1 {
		return nil, nil
	}
	return points[0], nil
}

// LatestNonZero returns the most recent non-zero datapoint for the given metric
func (s *Storage) LatestNonZero(metric string) (*datatypes.Datapoint, error) {
	if !ValidMetric(metric) {
		return nil, metricError(metric)
	}
	response, err := s.client.Query(client.Query{
		Command:  fmt.Sprintf("SELECT * FROM %s WHERE value != 0 ORDER BY DESC LIMIT 1", metric),
		Database: tableName,
	})
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}
	points, err := getDatapoints(metric, response)
	if err != nil {
		return nil, err
	}
	if len(points) != 1 {
		return nil, nil
	}
	return points[0], nil
}

// Close performs cleanup work
func (s *Storage) Close() error {
	return s.client.Close()
}

func metricError(metric string) error {
	return fmt.Errorf("illegal metric name: %v", metric)
}
