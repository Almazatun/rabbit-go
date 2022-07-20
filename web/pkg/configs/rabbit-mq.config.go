package configs

import "os"

var (
	RABBITMQ_DEFAULT_USER = os.Getenv("RABBITMQ_DEFAULT_USER")
	RABBITMQ_DEFAULT_PASS = os.Getenv("RABBITMQ_DEFAULT_PASS")
	MQ_HOST               = os.Getenv("MQ_HOST")
)
