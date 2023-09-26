package config

type Config struct {
	Env  string `json:"env"`
	Port int    `json:"port"`

	Kafka      Kafka    `json:"kafka"`
	ShovelList []Shovel `json:"shovelList"`
}

type Shovel struct {
	SourceTopicName      string `json:"sourceTopicName"`
	TargetTopicName      string `json:"targetTopicName"`
	PoisonedTopicName    string `json:"poisonedTopicName"`
	Name                 string `json:"name"`
	GroupId              string `json:"groupId"`
	AutoOffsetReset      string `json:"autoOffsetReset"`
	MaxErrorCount        int    `json:"errorCount"`
	MessageIntervalInSec int    `json:"messageIntervalInSec"`
}

type Kafka struct {
	BrokerList string `json:"brokerList"`
}
