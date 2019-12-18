package groupsync

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/event"
	"github.com/imulab/go-scim/protocol/groupsync"
	"github.com/imulab/go-scim/protocol/log"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"time"
)

// Create a new sender that listens to membership changes and publishes group sync messages.
func Sender(ch *amqp.Channel, logger log.Logger) (event.Publisher, error) {
	if err := declareRabbitQueue(ch, logger); err != nil {
		return nil, err
	}
	return &sender{
		ch:     ch,
		logger: logger,
	}, nil
}

type sender struct {
	ch     *amqp.Channel
	logger log.Logger
}

func (s *sender) ResourceCreated(ctx context.Context, created *prop.Resource) {
	go s.notify(created, groupsync.Compare(nil, created))
}

func (s *sender) ResourceUpdated(ctx context.Context, updated *prop.Resource, original *prop.Resource) {
	go s.notify(updated, groupsync.Compare(original, updated))
}

func (s *sender) ResourceDeleted(ctx context.Context, deleted *prop.Resource) {
	go s.notify(deleted, groupsync.Compare(deleted, nil))
}

func (s *sender) notify(group *prop.Resource, diff *groupsync.Diff) {
	if diff.CountJoined()+diff.CountLeft() == 0 {
		return
	}
	diff.ForEachJoined(func(id string) {
		s.submitMessage(group, id)
	})
	diff.ForEachLeft(func(id string) {
		s.submitMessage(group, id)
	})
}

func (s *sender) submitMessage(group *prop.Resource, memberId string) {
	msg := &message{
		GroupID:  group.ID(),
		MemberID: memberId,
		Trial:    1,
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error("failed to render message to json", log.Args{
			"error":    err,
			"groupId":  group.ID(),
			"memberId": memberId,
		})
		return
	}

	err = s.ch.Publish(
		exchangeName,
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			MessageId:   uuid.NewV4().String(),
			Timestamp:   time.Now(),
			Body:        raw,
		},
	)
	if err != nil {
		s.logger.Error("failed to send message to rabbit", log.Args{
			"error":    err,
			"message": fmt.Sprintf("%+v", msg),
		})
	}
}
