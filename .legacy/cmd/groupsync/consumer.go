package groupsync

import (
	"context"
	"encoding/json"
	job "github.com/imulab/go-scim/cmd/internal/groupsync"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/groupsync"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/service/filter"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"time"
)

type consumer struct {
	rabbitCh        *amqp.Channel
	userSyncService *groupsync.SyncService
	metaFilter      filter.ByResource
	userDatabase    db.DB
	groupDatabase   db.DB
	logger          *zerolog.Logger
	trialLimit      int
}

func (c *consumer) Start(ctx context.Context) (safeExit chan struct{}, err error) {
	messages, err := c.rabbitCh.Consume(
		job.RabbitQueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.logger.Err(err).Msg("failed to consume message")
		return
	}

	c.logger.Info().Msg("group sync consumer starts to listen for messages")

	safeExit = make(chan struct{}, 1)

	go func() {
		for message := range messages {
			select {
			case <-ctx.Done():
				safeExit <- struct{}{}
				return
			default:
				c.handle(message)
			}
		}
	}()

	return
}

func (c *consumer) handle(message amqp.Delivery) {
	payload := new(job.Message)
	if err := json.Unmarshal(message.Body, payload); err != nil {
		c.logger.Err(err).Fields(map[string]interface{}{
			"messageId": message.MessageId,
		}).Msg("Failed to unmarshal message payload")
		return
	}

	if payload.ExceededTrialLimit(c.trialLimit) {
		c.logger.
			Error().
			Fields(map[string]interface{}{
				"messageId":  message.MessageId,
				"trialLimit": c.trialLimit,
			}).
			Fields(payload.Fields()).
			Msg("Message had exceeded trial limit")
		return
	}

	isUser, err := c.assumeMemberIsUser(payload)
	if isUser {
		if err != nil {
			c.logger.Error().
				Fields(payload.Fields()).
				Fields(map[string]interface{}{"messageId": message.MessageId}).
				Msg("member is user but encountered error when syncing group property, will retry")
			c.retry(payload)
			return
		}

		c.logger.Info().
			Fields(payload.Fields()).
			Fields(map[string]interface{}{"messageId": message.MessageId}).
			Msg("member is user and was successfully synced")
		return
	}

	c.logger.Debug().
		Fields(payload.Fields()).
		Fields(map[string]interface{}{"messageId": message.MessageId}).
		Msg("member is not user, trying group")

	isGroup, err := c.assumeMemberIsGroup(payload)
	if isGroup {
		if err != nil {
			c.logger.Error().
				Fields(payload.Fields()).
				Fields(map[string]interface{}{"messageId": message.MessageId}).
				Msg("member is group but encountered error when expanding sync scope, will retry")
			c.retry(payload)
			return
		}

		c.logger.Info().
			Fields(payload.Fields()).
			Fields(map[string]interface{}{"messageId": message.MessageId}).
			Msg("member is group and sync scope was successfully expanded")
		return
	}

	c.logger.Error().
		Fields(payload.Fields()).
		Fields(map[string]interface{}{"messageId": message.MessageId}).
		Msg("member (possibly deleted) is neither user nor group, dropping message")
}

func (c *consumer) assumeMemberIsUser(payload *job.Message) (isUser bool, err error) {
	user, lookupErr := c.userDatabase.Get(context.Background(), payload.MemberID, nil)
	if lookupErr != nil || user == nil {
		return
	}

	isUser = true
	ref := user.Clone()

	err = c.userSyncService.SyncGroupPropertyForUser(context.Background(), user)
	if err != nil {
		return
	}

	if user.Hash() != ref.Hash() {
		err = c.metaFilter.FilterRef(context.Background(), user, ref)
		if err != nil {
			return
		}
		err = c.userDatabase.Replace(context.Background(), ref, user)
		if err != nil {
			return
		}
	}

	return
}

func (c *consumer) assumeMemberIsGroup(payload *job.Message) (isGroup bool, err error) {
	group, lookupErr := c.groupDatabase.Get(context.Background(), payload.MemberID, nil)
	if lookupErr != nil || group == nil {
		return
	}

	isGroup = true

	nav := group.Navigator().Dot("members")
	err = nav.Error()
	if err != nil {
		return
	}
	err = nav.ForEachChild(func(index int, child prop.Property) error {
		if value, err := child.ChildAtIndex("value"); err != nil {
			return err
		} else {
			c.send(&job.Message{
				GroupID:  payload.GroupID,
				MemberID: value.Raw().(string),
				Trial:    1,
			})
			return nil
		}
	})
	if err != nil {
		return
	}

	return
}

func (c *consumer) retry(message *job.Message) {
	message.Retry()
	c.send(message)
}

func (c *consumer) send(message *job.Message) {
	messageId := uuid.NewV4().String()

	raw, err := json.Marshal(message)
	if err != nil {
		c.logger.
			Err(err).
			Fields(map[string]interface{}{"messageId": messageId}).
			Fields(message.Fields()).
			Msg("Failed to send group sync message")
		return
	}

	err = c.rabbitCh.Publish(
		job.RabbitExchangeName,
		job.RabbitQueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			MessageId:   messageId,
			Timestamp:   time.Now(),
			Body:        raw,
		},
	)
	if err != nil {
		c.logger.
			Err(err).
			Fields(map[string]interface{}{"messageId": messageId}).
			Fields(message.Fields()).
			Msg("Failed to send group sync message")
		return
	}

	c.logger.
		Info().
		Fields(map[string]interface{}{"messageId": messageId}).
		Fields(message.Fields()).
		Msg("Sent group sync message")
}
