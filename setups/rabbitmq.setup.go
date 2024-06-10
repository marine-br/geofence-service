package setups

import (
	"github.com/marine-br/golib-rabbitmq/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"os"
)

func SetupRabbitmq() <-chan amqp091.Delivery {
	_, _, consumeChan := rabbitmq.ConsumeQueue(os.Getenv("RABBITMQ_CON_QUEUE"), 2500)

	return consumeChan
}
