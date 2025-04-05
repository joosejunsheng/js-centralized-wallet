package utils

import (
	"net/url"
	"strconv"
)

func GetQueryStr(q url.Values, key string, defaultValue string) string {
	value := q.Get(key)
	if value == "" {
		return defaultValue
	}

	return value
}

func GetQueryInt(q url.Values, key string, defaultValue int) int {
	value := q.Get(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}
