package merge

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"server/datatypes"
	"server/storage"
)

const (
	blockSize             = 1000
	timeout               = 3 * time.Second
	maxRetries            = 15
	timeFormatString      = "2006-01-02 15:04:05"
	defaultRemoteMergeURL = "https://solarracing.me/remotemerge"
)

var remoteMergeURL string

func init() {
	// Get the public URL of the remote server.
	url, ok := os.LookupEnv("REMOTE_SERVER_URL")
	if !ok {
		log.Println("Failed to find the public URL of the remote" +
			" server. Is the REMOTE_SERVER_URL environment variable set?")
		log.Printf("Defaulting to %s for REMOTE_SERVER_URL\n",
			defaultRemoteMergeURL)
		remoteMergeURL = defaultRemoteMergeURL
	} else {
		remoteMergeURL = strings.Trim(url, "\"")
	}
}

// Merger controls the logic for uploading data points from a local server to
// the remote server.
type Merger struct {
	model *Model
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
		store: store,
	}

	return merger, nil
}

// UploadLocalPointsToRemote finds all points (regardless of their metric)
// that were created between the specified time range, and begins the process
// of uploading those points to the server hosted at solarracing.me.
//
// This func is intended to run on a local server.
func (m *Merger) UploadLocalPointsToRemote(startTime, endTime time.Time) error {
	var curBlockNum int

	// Check the contents of merge_info_config.json to see if there is an
	// incomplete job for us to finish.
	if m.model.DidLastJobFinish == false {
		startTime = m.model.LastJobStartTimestamp
		endTime = m.model.LastJobEndTimestamp
		curBlockNum = m.model.LastJobBlockNumber

		msg := fmt.Sprintf("Resuming previous merge job that did not finish "+
			"(times %s to %s)",
			startTime.Format(timeFormatString),
			endTime.Format(timeFormatString),
		)
		log.Println(msg)
	} else {
		msg := fmt.Sprintf("Starting new merge job (times %s to %s)",
			startTime.Format(timeFormatString),
			endTime.Format(timeFormatString),
		)
		log.Println(msg)
	}

	// Get all points (of all metric types) within the current job's time range.
	metrics, err := m.store.ListMetrics()
	if err != nil {
		errMsg := "Failed to list all metrics in the data store." +
			" This shouldn't happen"
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	pointsToMerge := []*datatypes.Datapoint{}
	for _, metric := range metrics {
		newPoints, err := m.store.SelectMetricTimeRange(
			metric, startTime, endTime,
		)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to fetch points for the %s metric"+
				" within the current job's time range (times %s to %s): %v",
				metric,
				startTime.Format(timeFormatString),
				endTime.Format(timeFormatString),
				err.Error(),
			)
			log.Println(errMsg)
			return errors.New(errMsg)
		}

		pointsToMerge = append(pointsToMerge, newPoints...)
	}

	if len(pointsToMerge) <= 0 {
		errMsg := fmt.Sprintf("No points were collected locally within the"+
			" current job's time range (times %s to %s)",
			startTime.Format(timeFormatString),
			endTime.Format(timeFormatString),
		)
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	// Begin the process of merging the points in pointsToMerge in blocks.
	m.model.LastJobStartTimestamp = startTime
	m.model.LastJobEndTimestamp = endTime
	m.model.DidLastJobFinish = false
	m.model.Commit()

	msg := "Fetched points collected locally. Beginning the upload process" +
		" to the remote server..."
	log.Println(msg)

	c := make(chan bool, 1)
	retryCount := 0

	for ; curBlockNum <= len(pointsToMerge)/blockSize; curBlockNum++ {
		suffix := min(len(pointsToMerge), blockSize*(curBlockNum+1))
		if blockSize*curBlockNum < suffix {
			curBlock := pointsToMerge[blockSize*curBlockNum : suffix]
			curBlockAsJSON, err := json.Marshal(curBlock)
			if err != nil {
				return err
			}

			go func() {
				mergeBlockErr := mergeCurBlockOfPoints(curBlockAsJSON)
				if mergeBlockErr != nil {
					log.Println(mergeBlockErr.Error())
				}
				c <- mergeBlockErr == nil
			}()

			select {
			case res := <-c:
				if res {
					// Record that the current block was merged successfully.
					m.model.LastJobBlockNumber = curBlockNum
					m.model.Commit()
				} else {
					// Attempt to merge the current block again.
					curBlockNum--
					retryCount++
					if retryCount >= maxRetries {
						errMsg := "Max number of attempts to send blocks of" +
							" points exceeded. Aborting the current merge" +
							" operation and marking this job as incomplete..."
						log.Println(errMsg)
						return errors.New(errMsg)
					}

					msg := "Retrying to merge the last block of points..."
					log.Println(msg)
				}
			case <-time.After(timeout):
				// The current block took too long to merge. Attempt to merge
				// the current block again.
				curBlockNum--
				retryCount++
				if retryCount >= maxRetries {
					errMsg := "Max number of attempts to send blocks of" +
						" points exceeded. Aborting the current merge" +
						" operation and marking this job as incomplete..."
					log.Println(errMsg)
					return errors.New(errMsg)
				}
			}
		}
	}

	// Record that the current merge job finished successfully.
	m.model.DidLastJobFinish = true
	m.model.Commit()

	msg = fmt.Sprintf("Current merge job finished (times %s to %s)",
		startTime.Format(timeFormatString),
		endTime.Format(timeFormatString),
	)
	log.Println(msg)

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

// mergeCurBlockOfPoints fires a POST request to the /remotemerge endpoint
// that the remote server exposes with the provided block of points in the
// request body.
//
// If the request goes through with no problems, then this func pushes: true
// onto the provided channel; if the request doesn't go through for any
// reason, then this func pushes: false onto the provided channel.
func mergeCurBlockOfPoints(curBlockAsJSON []byte) error {
	// Hit the RemoteMergeHandler to merge the points into the remote
	// server's data store.
	res, err := http.Post(remoteMergeURL, "application/json", bytes.NewBuffer(curBlockAsJSON))
	if err != nil {
		return fmt.Errorf("Failed to send POST request to %s: %v",
			remoteMergeURL, err.Error())
	}
	if res.StatusCode != 204 {
		return fmt.Errorf("POST request to %s did not return 204: got: %d, want: 204",
			remoteMergeURL, res.StatusCode)
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
