package merge

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"server/datatypes"
	"server/message"
	"server/storage"
)

const (
	remoteMergeURL = "https://solarracing.me/remotemerge"
	blockSize      = 1000
)

// Merger controls the logic for uploading data points from a local server to
// the remote server.
type Merger struct {
	model *Model
	slack *message.SlackMessenger
	store *storage.Storage
}

// NewMerger returns a pointer to a new Merger object initialized with the
// provided values.
func NewMerger(store *storage.Storage) (*Merger, error) {
	model, err := ReadMergeInfoModel()
	if err != nil {
		return nil, err
	}

	merger := &Merger{
		model: model,
		slack: message.NewSlackMessenger(),
		store: store,
	}

	return merger, nil
}

// UploadLocalPointsToRemote finds all points (regardless of their metric)
// that were created between the specified time range, and begins the process
// of uploading those points to the server hosted at solarracing.me.
//
// This func is intended to run on a local server.
func (m *Merger) UploadLocalPointsToRemote(startTime, endTime *time.Time) error {
	// Get all points (of all metric types) within the specified time range.
	metrics, err := m.store.ListMetrics()
	if err != nil {
		errMsg := "Failed to list all metrics in the data store. This shouldn't happen"
		return errors.New(errMsg)
	}

	pointsToMerge := []*datatypes.Datapoint{}
	for _, metric := range metrics {
		newPoints, err := m.store.SelectMetricTimeRange(
			metric, *startTime, *endTime,
		)
		if err != nil {
			errMsg := "Failed to fetch points for the " + metric +
				" metric within the provided time range: " +
				err.Error()
			return errors.New(errMsg)
		}

		pointsToMerge = append(pointsToMerge, newPoints...)
	}

	if len(pointsToMerge) <= 0 {
		errMsg := "No points were collected locally within the specified time" +
			" range: " + startTime.Format("2006-01-02 15:04:05") + " to " +
			endTime.Format("2006-01-02 15:04:05")
		m.slack.PostNewMessage(errMsg)
		return errors.New(errMsg)
	}

	m.slack.PostNewMessage("Fetched points collected locally. Beginning the upload process to the remote server...")

	// Merge pointsToMerge in blocks
	for curBlockNum := 0; curBlockNum <= len(pointsToMerge)/blockSize; curBlockNum++ {
		suffix := min(len(pointsToMerge), blockSize*(curBlockNum+1))
		if blockSize*curBlockNum < suffix {
			// TODO: If we fail to merge the current block of points into the
			// remote server's data store, write the block number that failed
			// to merge_info_config.json, as well as the start and end
			// timestamps that all of the points were produced in between.
			//
			// Each time that this function is called, check if the block
			// number currently written to merge_info_config.json is equal to
			// 0. If it isn't, then resume that incomplete job by refetching
			// all of the points in between the saved start and end
			// timestamps, reconstruct the pointsToMerge array, and continue
			// the merging process at the saved block number.

			curBlock := pointsToMerge[blockSize*curBlockNum : suffix]

			curBlockAsJSON, err := json.Marshal(curBlock)
			if err != nil {
				return err
			}

			// Hit the RemoteMergeHandler to merge the points into the remote
			// server's data store.
			res, err := http.Post(remoteMergeURL, "application/json", bytes.NewBuffer(curBlockAsJSON))
			if err != nil {
				errMsg := "Failed to send POST request to " + remoteMergeURL +
					": " + err.Error()
				return errors.New(errMsg)
			}
			if res.StatusCode != 204 {
				errMsg := "POST request to " + remoteMergeURL +
					" did not return 204:" + err.Error()
				return errors.New(errMsg)
			}
		}
	}

	return nil
}

// MergePointsOntoRemote inserts the provided datapoints into the data store
// on the remote server.
//
// Despite this logic being really simple, it doesn't belong in the
// RemoteMergeHandler() func in api/merge.go. If this upload process ever
// needs to become more complicated, this is the place for that logic to be
// expanded upon.
//
// This func is intended to run on the remote server.
func (m *Merger) MergePointsOntoRemote(pointsToMerge []*datatypes.Datapoint) error {
	err := m.store.Insert(pointsToMerge)
	if err != nil {
		return err
	}
	return nil
}

// Helper func that should really be in the standard library if you ask me...
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
