// Home to constants for message classifiers. Used by our TCP messaging
// protocol to indicate what type of message is being sent (slack message, new
// track information, etc.)

package message

const (
	DatapointClassifier   = 'd'
	NumIncomingDatapoints = 'n'
)
