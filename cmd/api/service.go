package api

import (
	"context"
	"encoding/json"
	job "github.com/imulab/go-scim/cmd/internal/groupsync"
	"github.com/imulab/go-scim/pkg/v2/groupsync"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/service"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"time"
)

// groupCreated is a wrapper implementation of service.Create that computes the member joined the group and submit
// group property sync jobs for them.
type groupCreated struct {
	service service.Create
	sender  *groupSyncSender
}

func (s *groupCreated) Do(ctx context.Context, req *service.CreateRequest) (resp *service.CreateResponse, err error) {
	resp, err = s.service.Do(ctx, req)
	if err != nil {
		return
	}

	s.sender.Send(resp.Resource, groupsync.Compare(nil, resp.Resource))
	return
}

// groupReplaced is a wrapper implementation of service.Replace that computes the members joined and members left the
// group and submit group property sync jobs for them.
type groupReplaced struct {
	service service.Replace
	sender  *groupSyncSender
}

func (s *groupReplaced) Do(ctx context.Context, req *service.ReplaceRequest) (resp *service.ReplaceResponse, err error) {
	resp, err = s.service.Do(ctx, req)
	if err != nil || !resp.Replaced {
		return
	}

	s.sender.Send(resp.Resource, groupsync.Compare(resp.Ref, resp.Resource))
	return
}

// groupPatched is a wrapper implementation of service.Patch that computes the members joined and members left the
// group and submit group property sync jobs for them.
type groupPatched struct {
	service service.Patch
	sender  *groupSyncSender
}

func (s *groupPatched) Do(ctx context.Context, req *service.PatchRequest) (resp *service.PatchResponse, err error) {
	resp, err = s.service.Do(ctx, req)
	if err != nil || !resp.Patched {
		return
	}

	s.sender.Send(resp.Resource, groupsync.Compare(resp.Ref, resp.Resource))
	return
}

// groupDeleted is a wrapper implementation of service.Delete that computes the members left the group and submit group
// property sync jobs for them.
type groupDeleted struct {
	service service.Delete
	sender  *groupSyncSender
}

func (s *groupDeleted) Do(ctx context.Context, req *service.DeleteRequest) (resp *service.DeleteResponse, err error) {
	resp, err = s.service.Do(ctx, req)
	if err != nil {
		return
	}

	s.sender.Send(resp.Deleted, groupsync.Compare(resp.Deleted, nil))
	return
}

// groupSyncSender is an service that sends group sync messages for the groupsync.Diff object computed asynchronously
// to AMQP message brokers.
type groupSyncSender struct {
	channel *amqp.Channel
	logger  *zerolog.Logger
}

func (s *groupSyncSender) Send(group *prop.Resource, diff *groupsync.Diff) {
	if diff.CountLeft()+diff.CountJoined() == 0 {
		return
	}

	messageId := uuid.NewV4().String()
	s.logger.Info().Fields(map[string]interface{}{
		"messageId": messageId,
		"groupId":   group.IdOrEmpty(),
	}).Msg("Sending group sync messages.")

	go func(messageId string, diff *groupsync.Diff) {
		diff.ForEachLeft(func(id string) {
			s.submitMessage(messageId, group, id)
		})
		diff.ForEachJoined(func(id string) {
			s.submitMessage(messageId, group, id)
		})
	}(messageId, diff)
}

func (s *groupSyncSender) submitMessage(messageId string, group *prop.Resource, memberId string) {
	msg := job.Message{
		GroupID:  group.IdOrEmpty(),
		MemberID: memberId,
		Trial:    1,
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		s.logger.
			Err(err).
			Fields(map[string]interface{}{"messageId": messageId}).
			Fields(msg.Fields()).
			Msg("Failed to send group sync message")
		return
	}

	err = s.channel.Publish(
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
		s.logger.
			Err(err).
			Fields(map[string]interface{}{"messageId": messageId}).
			Fields(msg.Fields()).
			Msg("Failed to send group sync message")
		return
	}

	s.logger.
		Info().
		Fields(map[string]interface{}{"messageId": messageId}).
		Fields(msg.Fields()).
		Msg("Sent group sync message")
}
