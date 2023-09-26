package container

import (
	"github.com/burakolgun/gokafkashovel/producer"
	"github.com/burakolgun/gokafkashovel/resources/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

var cfg config.Config
var commonProducer *producer.CommonProducer
var logger zerolog.Logger
var redisClient *redis.Client

func GetCfg() *config.Config {
	return &cfg
}

func GetCommonProducer() *producer.CommonProducer {
	return commonProducer
}

func SetCommonProducer(_producer *producer.CommonProducer) {
	commonProducer = _producer
}

func GetLogger() *zerolog.Logger {
	return &logger
}

func SetLogger(_logger zerolog.Logger) {
	logger = _logger
}

func GetRedisClient() *redis.Client {
	return redisClient
}

func SetRedisClient(rcl *redis.Client) {
	redisClient = rcl
}
