package kafka_utils

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"strings"
)

func AddFieldToHeaderByFieldName(headers []kafka.Header, key string, value string) []kafka.Header {
	headers = append(headers, kafka.Header{Key: key, Value: []byte(value)})

	return headers
}

func GetFieldByNameFromHeader(headers []kafka.Header, fieldName string) string {
	for _, h := range headers {
		if strings.ToUpper(h.Key) == strings.ToUpper(fieldName) {
			return string(h.Value)
		}
	}

	return ""
}
