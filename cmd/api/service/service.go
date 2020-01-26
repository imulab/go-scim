package service

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/v2/pkg/groupsync"
	"github.com/imulab/go-scim/v2/pkg/prop"
	"github.com/imulab/go-scim/v2/pkg/service"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"time"
)

// GroupCreated is a wrapper implementation of service.Create that computes the member joined the group and submit
// group property sync jobs for them.
type GroupCreated struct {
	Service service.Create
	Sender  *GroupSyncSender
}

func (s *GroupCreated) Do(ctx context.Context, req *service.CreateRequest) (resp *service.CreateResponse, err error) {
	resp, err = s.Service.Do(ctx, req)
	if err != nil {
		return
	}

	s.Sender.Send(resp.Resource, groupsync.Compare(nil, resp.Resource))
	return
}

// GroupReplaced is a wrapper implementation of service.Replace that computes the members joined and members left the
// group and submit group property sync jobs for them.
type GroupReplaced struct {
	Service service.Replace
	Sender  *GroupSyncSender
}

func (s *GroupReplaced) Do(ctx context.Context, req *service.ReplaceRequest) (resp *service.ReplaceResponse, err error) {
	resp, err = s.Service.Do(ctx, req)
	if err != nil || !resp.Replaced {
		return
	}

	s.Sender.Send(resp.Resource, groupsync.Compare(resp.Ref, resp.Resource))
	return
}

// GroupPatched is a wrapper implementation of service.Patch that computes the members joined and members left the
// group and submit group property sync jobs for them.
type GroupPatched struct {
	Service service.Patch
	Sender  *GroupSyncSender
}

func (s *GroupPatched) Do(ctx context.Context, req *service.PatchRequest) (resp *service.PatchResponse, err error) {
	resp, err = s.Service.Do(ctx, req)
	if err != nil || !resp.Patched {
		return
	}

	s.Sender.Send(resp.Resource, groupsync.Compare(resp.Ref, resp.Resource))
	return
}

// GroupDeleted is a wrapper implementation of service.Delete that computes the members left the group and submit group
// property sync jobs for them.
type GroupDeleted struct {
	Service service.Delete
	Sender  *GroupSyncSender
}

func (s *GroupDeleted) Do(ctx context.Context, req *service.DeleteRequest) (resp *service.DeleteResponse, err error) {
	resp, err = s.Service.Do(ctx, req)
	if err != nil {
		return
	}

	s.Sender.Send(resp.Deleted, groupsync.Compare(resp.Deleted, nil))
	return
}

// GroupSyncSender is an service that sends group sync messages for the groupsync.Diff object computed asynchronously
// to AMQP message brokers.
type GroupSyncSender struct {
	Channel *amqp.Channel
	Logger  *zerolog.Logger
}

func (s *GroupSyncSender) Send(group *prop.Resource, diff *groupsync.Diff) {
	if diff.CountLeft()+diff.CountJoined() == 0 {
		return
	}

	messageId := uuid.NewV4().String()
	s.Logger.Info().Fields(map[string]interface{}{
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

func (s *GroupSyncSender) submitMessage(messageId string, group *prop.Resource, memberId string) {
	msg := struct {
		GroupID  string `json:"group_id"`
		MemberID string `json:"member_id"`
		Trial    int    `json:"trial"`
	}{
		GroupID:  group.IdOrEmpty(),
		MemberID: memberId,
		Trial:    1,
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		s.Logger.Err(err).Fields(map[string]interface{}{
			"messageId": messageId,
			"groupId":   group.IdOrEmpty(),
			"memberId":  memberId,
		}).Msg("Failed to send group sync message")
		return
	}

	err = s.Channel.Publish(
		"",
		"group_sync",
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
		s.Logger.Err(err).Fields(map[string]interface{}{
			"messageId": messageId,
			"groupId":   group.IdOrEmpty(),
			"memberId":  memberId,
		}).Msg("Failed to send group sync message")
		return
	}

	s.Logger.Info().Fields(map[string]interface{}{
		"messageId": messageId,
		"groupId":   group.IdOrEmpty(),
		"memberId":  memberId,
	}).Msg("Sent group sync message")
}
