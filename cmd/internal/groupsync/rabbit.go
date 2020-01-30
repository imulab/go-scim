package groupsync

import "github.com/streadway/amqp"

// Queue name in RabbitMQ, used by consumer and producer to exchange message.
const RabbitQueueName = "group_sync"

// Exchange name in RabbitMQ
const RabbitExchangeName = ""

// Declare a queue in RabbitMQ named RabbitQueueName.
func DeclareQueue(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(
		RabbitQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}
