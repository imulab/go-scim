package groupsync

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/groupsync"
	"github.com/imulab/go-scim/protocol/log"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"time"
)

// Start the message consuming process with the cancellable context and return a safe exit channel. The cancellable
// context can be used to notify the consumer that it should abort. However, the consumer only checks for this signal
// before processing each message and hence may not immediately abort. The returned safe exit channel can be received
// by the caller process when the consumer is truly ready to exit.
func StartConsumer(ctx context.Context, appCtx *appContext, args *args) (safeExit chan struct{}, err error) {
	if err := declareRabbitQueue(appCtx.rabbitCh, appCtx.logger); err != nil {
		return nil, err
	}
	c := &consumer{
		ch:         appCtx.rabbitCh,
		userDB:     appCtx.userDB,
		groupDB:    appCtx.groupDB,
		trialLimit: args.requeueLimit,
		logger:     appCtx.logger,
	}
	safeExit, err = c.Consume(ctx)
	return
}

type consumer struct {
	ch         *amqp.Channel
	userDB     db.DB
	groupDB    db.DB
	trialLimit int
	logger     log.Logger
}

func (c *consumer) Consume(ctx context.Context) (chan struct{}, error) {
	msgs, err := c.ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.logger.Error("failed to consume message", log.Args{
			"error": err,
		})
		return nil, err
	}

	exitChan := make(chan struct{}, 1)

	go func() {
		for msg := range msgs {
			select {
			case <-ctx.Done():
				exitChan <- struct{}{}
				return
			default:
				c.handle(msg)
			}
		}
	}()

	return exitChan, nil
}

func (c *consumer) handle(msg amqp.Delivery) {
	message := new(message)
	if err := json.Unmarshal(msg.Body, message); err != nil {
		c.logger.Error("message body is corrupted, will drop message", log.Args{
			"error":   err,
			"message": string(msg.Body),
		})
		return
	}

	if c.trialLimit > 0 && message.Trial > c.trialLimit {
		c.logger.Error("message had exceeded trial limit, will drop message", log.Args{
			"message": string(msg.Body),
			"trial":   message.Trial,
			"limit":   c.trialLimit,
		})
		return
	}

	if user, err := c.userDB.Get(context.Background(), message.MemberID, nil); err == nil && user != nil {
		if err := c.syncUserGroup(user); err != nil {
			c.logger.Error("error encountered while syncing user group, will requeue", log.Args{
				"error":   err,
				"message": fmt.Sprintf("%+v", message),
			})
			c.send(message)
		} else {
			c.logger.Info("successfully synced user group", log.Args{
				"userId": message.MemberID,
			})
		}
		return
	}

	c.logger.Debug("message does not entail user, assuming group resource", log.Args{
		"message": fmt.Sprintf("%+v", message),
	})

	if group, err := c.groupDB.Get(context.Background(), message.MemberID, nil); err == nil && group != nil {
		if err := c.expandSyncScope(message, group); err != nil {
			c.logger.Error("error encountered while expanding sync scope for group member, will requeue", log.Args{
				"error":   err,
				"message": fmt.Sprintf("%+v", message),
			})
			c.send(message)
		} else {
			c.logger.Info("successfully expand sync scope for group", log.Args{
				"groupId": message.MemberID,
			})
		}
		return
	}

	c.logger.Error("message entail neither user nor group, aborted", log.Args{
		"message": fmt.Sprintf("%+v", message),
	})
}

func (c *consumer) syncUserGroup(user *prop.Resource) error {
	if err := groupsync.Refresher(c.groupDB).Refresh(context.Background(), user); err != nil {
		return err
	}
	if err := c.userDB.Replace(context.Background(), user); err != nil {
		return err
	}
	return nil
}

func (c *consumer) expandSyncScope(originalMsg *message, group *prop.Resource) error {
	if nav := group.NewFluentNavigator().FocusName("members"); nav.Error() != nil {
		return nav.Error()
	} else {
		_ = nav.CurrentAsContainer().ForEachChild(func(index int, child prop.Property) error {
			if vp := child.(prop.Container).ChildAtIndex("value"); vp != nil && !vp.IsUnassigned() {
				c.send(&message{
					GroupID:  originalMsg.GroupID,
					MemberID: vp.Raw().(string),
				})
			}
			return nil
		})
		return nil
	}
}

func (c *consumer) send(message *message) {
	message.Trial++

	action := "send"
	if message.Trial > 1 {
		action = "requeue"
	}

	raw, err := json.Marshal(message)
	if err != nil {
		c.logger.Error(fmt.Sprintf("failed to parse message to JSON, will not %s", action), log.Args{
			"error":   err,
			"message": fmt.Sprintf("%+v", message),
		})
		return
	}
	err = c.ch.Publish(
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
		c.logger.Error(fmt.Sprintf("failed to %s message to rabbit", action), log.Args{
			"error":   err,
			"message": fmt.Sprintf("%+v", message),
		})
	}
}
