package groupsync

import (
	"fmt"
	"github.com/imulab/go-scim/protocol/log"
)

const subject = "group_diff"

// Message sent to notify the sync worker to sync the user resource with MemberID, or expand
// the group resource with MemberID into more sync tasks.
type message struct {
	GroupID  string `json:"group_id"`
	MemberID string `json:"member_id"`
	Trial    int    `json:"trial"`
}

func (m *message) logSent(logger log.Logger) {
	logger.Info("sent %s message for group [id=%s] and member [id=%s]", subject, m.GroupID, m.MemberID)
}

func (m *message) logReturned(logger log.Logger) {
	logger.Info("returned %s message for group [id=%s] and member [id=%s]", subject, m.GroupID, m.MemberID)
}

func (m *message) logFailed(logger log.Logger) {
	logger.Error("failed to send %s message for group [id=%s] and member [id=%s]", subject, m.GroupID, m.MemberID)
}

func (m *message) String() string {
	return fmt.Sprintf("[groupId=%s memberId=%s]", m.GroupID, m.MemberID)
}
