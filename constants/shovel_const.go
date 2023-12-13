package constants

const (
	RequestIdKey                         = "requestID"
	KafkaErrorCountKey                   = "x-error-count"
	KafkaErrorCountDefaultValue          = 1
	KafkaPoisonedTopicCycleCountKey      = "x-poisoned-cycle-count"
	KafkaPosisonedTopicCycleDefaultValue = -1

	ConsumerClosed = "consumer is closed"
)
