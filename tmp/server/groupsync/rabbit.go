package groupsync

import (
	"github.com/imulab/go-scim/protocol/log"
	"github.com/streadway/amqp"
)

const (
	exchangeName = ""
	queueName    = "group_sync"
)

// Declare a durable and non-autoDelete RabbitMQ queue with the name "group_sync".
func declareRabbitQueue(ch *amqp.Channel, logger log.Logger) error {
	_, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error("failed to declare rabbit queue", log.Args{
			"error": err,
		})
	}
	return nil
}

// Message sent to notify the sync worker to sync the user resource with MemberID, or expand
// the group resource with MemberID into more sync tasks.
type message struct {
	GroupID  string `json:"group_id"`
	MemberID string `json:"member_id"`
	Trial    int    `json:"trial"`
}
