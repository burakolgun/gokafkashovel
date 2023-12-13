package config

type Config struct {
	Env  string `json:"env"`
	Port int    `json:"port"`

	Kafka                      Kafka    `json:"kafka"`
	ShovelList                 []Shovel `json:"shovelList"`
	ShovelIntervalInSec        int      `json:"messageIntervalInSec"`
	ShovelWaitingIntervalInSec int      `json:"shovelWaitingIntervalInSec"`
}

type Shovel struct {
	SourceTopicName          string `json:"sourceTopicName"`
	TargetTopicName          string `json:"targetTopicName"`
	PoisonedTopicName        string `json:"poisonedTopicName"`
	Name                     string `json:"name"`
	GroupId                  string `json:"groupId"`
	AutoOffsetReset          string `json:"autoOffsetReset"`
	MaxErrorCount            int    `json:"errorCount"`
	IsPoisonedTopicCycleOpen bool   `json:"isPoisonedTopicCycleOpen"`
}

type Kafka struct {
	BrokerList string `json:"brokerList"`
}
