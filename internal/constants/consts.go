package constants

import (
	"os"
)

var (
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_HOST     string
	POSTGRES_DB       string

	REDIS_HOST string
	REDIS_PORT string
)

func init() {
	POSTGRES_USER = os.Getenv("POSTGRES_USER")
	POSTGRES_PASSWORD = os.Getenv("POSTGRES_PASSWORD")
	POSTGRES_HOST = os.Getenv("POSTGRES_HOST")
	POSTGRES_DB = os.Getenv("POSTGRES_DB")

	REDIS_HOST = os.Getenv("REDIS_HOST")
	REDIS_PORT = os.Getenv("REDIS_PORT")
}
