// Home to constants for message classifiers. Used by our TCP messaging
// protocol to indicate what type of message is being sent (slack message, new
// track information, etc.)

package message

const (
	SlackMessageClassifier          = 'c'
	DataPointClassifier             = 'd'
	NumIncomingDataPointsClassifier = 't' //consider re-naming for clarity
)
