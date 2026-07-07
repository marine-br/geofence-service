package setups

import (
	"os"
	"strconv"

	"github.com/marine-br/golib-rabbitmq/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

func SetupRabbitmq() <-chan amqp091.Delivery {
	prefetch := 100
	if v := os.Getenv("RABBITMQ_PREFETCH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			prefetch = n
		}
	}

	_, _, consumeChan := rabbitmq.ConsumeQueue(os.Getenv("RABBITMQ_CON_QUEUE"), prefetch)

	return consumeChan
}
